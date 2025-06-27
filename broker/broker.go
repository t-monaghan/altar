// Package broker provides a broker for the Awtrix firmware
//
// This broker calls the altar application fetchers, and posts the updated data to the Awtrix host
package broker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/notifier"
	"github.com/t-monaghan/altar/utils"
	"github.com/t-monaghan/altar/utils/awtrix"
)

const httpTimeout = 10 * time.Second
const idleTimeout = 120 * time.Second
const mockAwtrixPort = ":8080"

// DefaultAdminPort is the default port for the broker's api.
const DefaultAdminPort = "25827"

// AltarAdminRequest defines the expected request type for the altar admin server.
type AltarAdminRequest struct {
	Command AltarAdminCommand `json:"command"`
	Data    string            `json:"data,omitempty"`
}

// AltarAdminCommand defines the commands the altar admin server recognises.
type AltarAdminCommand string

const (
	// AdminShutdownCommand is the command recognised by altar's admin server as a call to shutdown.
	AdminShutdownCommand AltarAdminCommand = "DOWN"
)

// HTTPBroker performs each routine's fetching, hosts handler functions on it's server and communicates updates to
// the Awtrix host.
type HTTPBroker struct {
	routines      []utils.Routine
	clockAddress  string
	Client        *http.Client
	DebugMode     bool
	MockAwtrix    bool
	DisplayConfig awtrix.Config
	AdminPort     string
	handlers      map[string]func(http.ResponseWriter, *http.Request)
}

// ErrBrokerHasNoApplications occurs when an altar Broker is instantiated with no applications.
var ErrBrokerHasNoApplications = errors.New("failed to initialise broker: no applications were provided")

// ErrIPNotValid occurs when an invalid IP address is provided.
var ErrIPNotValid = errors.New("IP address is not valid")

// NewBroker instantiates a new altar broker.
func NewBroker(
	awtrixAddress string,
	routines []utils.Routine,
	handlers map[string]func(http.ResponseWriter, *http.Request),
	options ...func(*awtrix.Config),
) (*HTTPBroker, error) {
	if len(routines) == 0 {
		return nil, ErrBrokerHasNoApplications
	}

	clockIP := net.ParseIP(awtrixAddress)
	if clockIP == nil {
		return nil, ErrIPNotValid
	}

	cfg := awtrix.Config{}
	for _, option := range options {
		option(&cfg)
	}

	brkr := HTTPBroker{
		clockAddress:  fmt.Sprintf("http://%v", clockIP),
		routines:      routines,
		Client:        &http.Client{Timeout: httpTimeout},
		DebugMode:     false,
		DisplayConfig: cfg,
		handlers:      handlers,
	}

	return &brkr, nil
}

// Start begins execution of the broker's routine.
func (b *HTTPBroker) Start() {
	if b.DebugMode {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		// avoid rebooting when debugging
		err := b.sendConfig()
		if err != nil {
			slog.Error("error setting up initial awtrix configuration", "error", err)
		}

		slog.Info("rebooting awtrix device")

		// rebooting awtrix is required to ensure the configuration is applied
		err = b.rebootAwtrix()
		if err != nil {
			slog.Error("error rebooting the awtrix device", "error", err)
		}
	}

	go func() {
		fetchAndPushApps(b)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/admin/command", commandHandler)

	for path, handler := range b.handlers {
		mux.HandleFunc(path, handler)
	}

	adminPort := DefaultAdminPort

	if b.AdminPort != "" {
		adminPort = b.AdminPort
	}

	adminServer := &http.Server{
		Addr:         ":" + adminPort,
		Handler:      mux,
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Fatal(adminServer.ListenAndServe())
}

func fetchAndPushApps(brkr *HTTPBroker) {
	for {
		startTime := time.Now()

		var quickestPoll = time.Hour * 9000

		var fetchGroup sync.WaitGroup

		var mutateConfigAndSetPollRate sync.Mutex

		for _, app := range brkr.routines {
			fetchGroup.Add(1)

			go func(app utils.Routine) {
				defer fetchGroup.Done()

				defer func() {
					if r := recover(); r != nil {
						slog.Error("broker has recovered from fetcher panicking", "error", r)
					}
				}()

				err := app.Fetch(brkr.Client)
				if err != nil {
					slog.Error("error encountered in fetching", "app", app.GetName(), "error", err)
				}

				mutateConfigAndSetPollRate.Lock()
				brkr.DisplayConfig = mergeConfig(brkr.DisplayConfig, app.GetGlobalConfig())

				if app.GetPollRate() < quickestPoll {
					quickestPoll = app.GetPollRate()
				}
				mutateConfigAndSetPollRate.Unlock()
			}(app)
		}

		fetchGroup.Wait()

		err := brkr.sendConfig()
		if err != nil {
			slog.Error("error changing awtrix settings", "error", err)
		}

		for _, app := range brkr.routines {
			err := brkr.push(app)
			if err != nil {
				slog.Error("error encountered pushing to awtrix device", "app", app.GetName(), "error", err)
			}
		}

		duration := time.Since(startTime)
		if duration < quickestPoll {
			time.Sleep(quickestPoll - duration)
		}
	}
}

func (b *HTTPBroker) sendConfig() error {
	if b.DisplayConfig.TransitionEffect != nil {
		next := *b.DisplayConfig.TransitionEffect + 1
		b.DisplayConfig.TransitionEffect = &next
	} else {
		zero := 0
		b.DisplayConfig.TransitionEffect = &zero
	}

	jsonData, err := json.Marshal(b.DisplayConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal awtrix config into json: %w", err)
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	port := ""
	if b.MockAwtrix {
		port = mockAwtrixPort
	}

	address := fmt.Sprintf("%v%v/api/settings", b.clockAddress, port)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, address, bufferedJSON)
	if err != nil {
		return fmt.Errorf("failed to create post request for awtrix configuration: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform post request for awtrix configuration: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	if utils.ResponseStatusIsNot2xx(resp.StatusCode) {
		slog.Error("awtrix has responded to configuration request with non-2xx http response", "http-status", resp.Status)
	}

	return err
}

// ErrUnknownRoutineType is thrown when altar encounters a concrete routine type that it does not recognise.
// Strictly only the types defined by altar are used as the broker's push method needs to know how to handle the
// request.
var ErrUnknownRoutineType = errors.New("unknown routine type")

func (b *HTTPBroker) push(routine utils.Routine) error {
	if !routine.ShouldPushToAwtrix() {
		slog.Debug("skipping push for routine", "routine", routine.GetName())

		return nil
	}

	jsonData, err := json.Marshal(routine.GetData())
	if err != nil {
		return fmt.Errorf("failed to marshal %v data into json: %w", routine.GetName(), err)
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	port := ""
	if b.MockAwtrix {
		port = mockAwtrixPort
	}

	var address string
	switch routine.(type) {
	case *application.Application:
		address = fmt.Sprintf("%v%v/api/custom?name=%v", b.clockAddress, port, url.QueryEscape(routine.GetName()))
	case *notifier.Notifier:
		address = fmt.Sprintf("%v%v/api/notify", b.clockAddress, port)
	default:
		return fmt.Errorf("%w for routine: %v", ErrUnknownRoutineType, routine.GetName())
	}

	err = b.postRequestToAwtrix(address, bufferedJSON, routine.GetName())
	if err != nil {
		return err
	}

	slog.Debug("pushed", "routine-name", routine.GetName())

	return err
}

func (b *HTTPBroker) postRequestToAwtrix(address string, bufferedJSON *bytes.Buffer, routineName string) error {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, address, bufferedJSON)
	if err != nil {
		return fmt.Errorf("failed to create post request for %v: %w", routineName, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform post request for %v: %w", routineName, err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("%w for routine %v: %w", utils.ErrClosingResponseBody, routineName, closeErr)
		}
	}()

	if utils.ResponseStatusIsNot2xx(resp.StatusCode) {
		slog.Error("awtrix has responded to push with non-2xx http response",
			"http-status", resp.Status, "routine", routineName)
	}

	return err
}

func commandHandler(wrtr http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(wrtr, "Error reading request body", http.StatusInternalServerError)

		return
	}

	err = req.Body.Close()
	if err != nil {
		slog.Error("error in shutdown handler", "error", err)
	}

	if req.Method != http.MethodPost {
		wrtr.WriteHeader(http.StatusBadRequest)
		_, _ = wrtr.Write([]byte("request to admin commands did not use the POST method"))

		return
	}

	requestCommand := &AltarAdminRequest{}

	err = json.Unmarshal(body, requestCommand)
	if err != nil {
		slog.Error("admin server failed to unmarshal command request")
		wrtr.WriteHeader(http.StatusBadRequest)

		return
	}

	switch requestCommand.Command {
	case AdminShutdownCommand:
		slog.Info("admin server received shutdown command - shutting down")
		os.Exit(0)
	default:
		wrtr.WriteHeader(http.StatusBadRequest)
		_, _ = wrtr.Write([]byte("admin server did not recognise the command: '" + string(body) + "'"))

		return
	}
}

func (b *HTTPBroker) rebootAwtrix() error {
	port := ""
	if b.MockAwtrix {
		port = mockAwtrixPort
	}

	address := fmt.Sprintf("%v%v/api/reboot", b.clockAddress, port)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, address, nil)
	if err != nil {
		return fmt.Errorf("failed to create post request for rebooting awtrix device: %w", err)
	}

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform post request for rebooting awtrix device: %w", err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	if utils.ResponseStatusIsNot2xx(resp.StatusCode) {
		slog.Error("awtrix has responded to reboot command with non-2xx http response",
			"http-status", resp.Status)
	}

	return err
}

func mergeConfig(left awtrix.Config, right awtrix.Config) awtrix.Config {
	keep := awtrix.Config{}
	if right.Overlay != "" {
		keep.Overlay = right.Overlay
	} else if left.Overlay != "" {
		keep.Overlay = left.Overlay
	}

	return keep
}
