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
)

// MinLoopTime is the minimum time the broker will spend between iterations of fetching and pushing updates.
const MinLoopTime = 5 * time.Second
const httpTimeout = 10 * time.Second
const idleTimeout = 120 * time.Second
const defaultWebPort = ":8080"

// DefaultAdminPort is the port the broker listens on for commands.
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

// HTTPBroker is a broker that queries each of the Altar applications and communicates updates to the Awtrix host.
type HTTPBroker struct {
	// applications  []*application.Application
	// notifiers     []*notifier.Notifier
	handlers      []utils.AltarHandler
	clockAddress  string
	Client        *http.Client
	Debug         bool
	DisplayConfig utils.AwtrixConfig
	AdminPort     string
}

// ErrBrokerHasNoApplications occurs when an Altar Broker is instantiated with no applications.
var ErrBrokerHasNoApplications = errors.New("failed to initialise broker: no applications were provided")

// ErrIPNotValid occurs when an Altar Broker is instantiated with an invalid IP address.
var ErrIPNotValid = errors.New("failed to initialise broker: IP address is not valid")

// NewBroker instantiates a new Altar broker.
func NewBroker(
	addr string,
	applications []utils.AltarHandler,
	options ...func(*utils.AwtrixConfig),
) (*HTTPBroker, error) {
	if len(applications) == 0 {
		return nil, ErrBrokerHasNoApplications
	}

	clockIP := net.ParseIP(addr)
	if clockIP == nil {
		return nil, ErrIPNotValid
	}

	cfg := utils.AwtrixConfig{}
	for _, option := range options {
		option(&cfg)
	}

	brkr := HTTPBroker{
		clockAddress:  fmt.Sprintf("http://%v", clockIP),
		handlers:      applications,
		Client:        &http.Client{Timeout: httpTimeout},
		Debug:         false,
		DisplayConfig: cfg,
	}

	return &brkr, nil
}

// Start executes the broker's routine.
func (b *HTTPBroker) Start() {
	if b.Debug {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

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

	go func() {
		fetchAndPushApps(b)
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/admin/command", commandHandler)

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

		var mutateConfig sync.Mutex

		for _, app := range brkr.handlers {
			fetchGroup.Add(1)

			go func(app utils.AltarHandler) {
				defer fetchGroup.Done()

				err := app.Fetch(brkr.Client)
				if err != nil {
					slog.Error("error encountered in fetching", "app", app.GetName(), "error", err)
				}

				mutateConfig.Lock()
				if app.GetPollRate() < quickestPoll {
					brkr.DisplayConfig = mergeConfig(brkr.DisplayConfig, app.GetGlobalConfig())
					quickestPoll = app.GetPollRate()
				}
				mutateConfig.Unlock()
			}(app)
		}

		fetchGroup.Wait()

		err := brkr.sendConfig()
		if err != nil {
			slog.Error("error changing awtrix settings", "error", err)
		}

		// TODO: decide if push should be sent as batch request, or similarly to above with goroutines
		// https://github.com/Blueforcer/awtrix3/blob/main/docs/api.md#sending-multiple-custom-apps-simultaneously
		// TODO: decide if apps should be pushed as soon as they're fetched
		for _, app := range brkr.handlers {
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
	jsonData, err := json.Marshal(b.DisplayConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal awtrix config into json: %w", err)
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	debugPort := ""
	if b.Debug {
		debugPort = defaultWebPort
	}

	address := fmt.Sprintf("%v%v/api/settings", b.clockAddress, debugPort)

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

// ErrUnknownHandlerType is thrown when altar encounters a concrete handler type that it does not recognise.
// Strictly only the types defined by altar are used as the broker's push method needs to know how to handle the
// request.
var ErrUnknownHandlerType = errors.New("unknown handler type")

//nolint:cyclop
func (b *HTTPBroker) push(handler utils.AltarHandler) error {
	if !handler.ShouldPushToAwtrix() {
		slog.Debug("skipping push for handler", "handler", handler.GetName())

		return nil
	}

	jsonData, err := json.Marshal(handler.GetData())
	if err != nil {
		return fmt.Errorf("failed to marshal %v data into json: %w", handler.GetName(), err)
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	debugPort := ""
	if b.Debug {
		debugPort = defaultWebPort
	}

	var address string
	switch handler.(type) {
	case *application.Application:
		address = fmt.Sprintf("%v%v/api/custom?name=%v", b.clockAddress, debugPort, url.QueryEscape(handler.GetName()))
	case *notifier.Notifier:
		address = fmt.Sprintf("%v%v/api/notify", b.clockAddress, debugPort)
	default:
		return fmt.Errorf("%w for handler: %v", ErrUnknownHandlerType, handler.GetName())
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, address, bufferedJSON)
	if err != nil {
		return fmt.Errorf("failed to create post request for %v: %w", handler.GetName(), err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform post request for %v: %w", handler.GetName(), err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("%w for handler %v: %w", utils.ErrClosingResponseBody, handler.GetName(), closeErr)
		}
	}()

	if utils.ResponseStatusIsNot2xx(resp.StatusCode) {
		slog.Error("awtrix has responded to push with non-2xx http response",
			"http-status", resp.Status, "handler", handler.GetName())
	}

	slog.Debug("pushed", "handler-name", handler.GetName())

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
	debugPort := ""
	if b.Debug {
		debugPort = defaultWebPort
	}

	address := fmt.Sprintf("%v%v/api/reboot", b.clockAddress, debugPort)

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

// TODO: remove this, have apps return a method which is performed on the config a la the blog post
// mergeConfig will only merge options considered worth changing at runtime.
func mergeConfig(left utils.AwtrixConfig, right utils.AwtrixConfig) utils.AwtrixConfig {
	keep := utils.AwtrixConfig{}
	if right.Overlay != "" {
		keep.Overlay = right.Overlay
	} else if left.Overlay != "" {
		keep.Overlay = left.Overlay
	}

	return keep
}
