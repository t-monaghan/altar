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
	"time"

	"github.com/t-monaghan/altar/application"
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
	applications  []*application.Application
	clockAddress  string
	Client        *http.Client
	Debug         bool
	DisplayConfig AwtrixConfig
	AdminPort     string
}

// ErrBrokerHasNoApplications occurs when an Altar Broker is instantiated with no applications.
var ErrBrokerHasNoApplications = errors.New("failed to initialise broker: no applications were provided")

// ErrIPNotValid occurs when an Altar Broker is instantiated with an invalid IP address.
var ErrIPNotValid = errors.New("failed to initialise broker: IP address is not valid")

// NewBroker instantiates a new Altar broker.
func NewBroker(
	addr string,
	applications []*application.Application,
	options ...func(*AwtrixConfig),
) (*HTTPBroker, error) {
	if len(applications) == 0 {
		return nil, ErrBrokerHasNoApplications
	}

	clockIP := net.ParseIP(addr)
	if clockIP == nil {
		return nil, ErrIPNotValid
	}

	cfg := AwtrixConfig{}
	for _, option := range options {
		option(&cfg)
	}

	brkr := HTTPBroker{
		clockAddress:  fmt.Sprintf("http://%v", clockIP),
		applications:  applications,
		Client:        &http.Client{Timeout: httpTimeout},
		Debug:         false,
		DisplayConfig: cfg,
	}

	return &brkr, nil
}

// Start executes the broker's routine.
func (b *HTTPBroker) Start() {
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
		for {
			startTime := time.Now()

			var quickestPoll = time.Hour * 9000

			for _, app := range b.applications {
				err := app.Fetch()
				if err != nil {
					slog.Error("error fetching %v: %v", app.Name, err)
				}

				if app.PollRate < quickestPoll {
					quickestPoll = app.PollRate
				}
			}

			for _, app := range b.applications {
				err := b.push(app)
				if err != nil {
					slog.Error("error pushing %v: %v", app.Name, err)
				}
			}

			duration := time.Since(startTime)
			if duration < quickestPoll {
				time.Sleep(quickestPoll - duration)
			}
		}
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

func (b *HTTPBroker) push(app *application.Application) error {
	if !app.ShouldPushToAwtrix() {
		slog.Debug("skipping push for app", "app", app.Name)

		return nil
	}

	app.HasUnpushedData = false

	jsonData, err := json.Marshal(app.GetData())
	if err != nil {
		return fmt.Errorf("failed to marshal %v data into json: %w", app.Name, err)
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	debugPort := ""
	if b.Debug {
		debugPort = defaultWebPort
	}

	address := fmt.Sprintf("%v%v/api/custom?name=%v", b.clockAddress, debugPort, url.QueryEscape(app.Name))

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, address, bufferedJSON)
	if err != nil {
		return fmt.Errorf("failed to create post request for %v: %w", app.Name, err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform post request for %v: %w", app.Name, err)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = fmt.Errorf("failed to close body of app push request: %w", closeErr)
		}
	}()

	if utils.ResponseStatusIsNot2xx(resp.StatusCode) {
		slog.Error("awtrix has responded to app update with non-2xx http response",
			"http-status", resp.Status, "app", app.Name)
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
