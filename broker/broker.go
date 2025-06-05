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
)

// MinLoopTime is the minimum time the broker will spend between iterations of fetching and pushing updates.
const MinLoopTime = 5 * time.Second
const httpTimeout = 10 * time.Second
const idleTimeout = 120 * time.Second

// AdminPort is the port the broker listens on for commands.
const AdminPort = "25827"

// HTTPBroker is a broker that queries each of the Altar applications and communicates updates to the Awtrix host.
type HTTPBroker struct {
	applications  []*application.Application
	clockAddress  string
	Client        *http.Client
	Debug         bool
	DisplayConfig awtrixConfig
}

// ErrBrokerHasNoApplications occurs when an Altar Broker is instantiated with no applications.
var ErrBrokerHasNoApplications = errors.New("failed to initialise broker: no applications were provided")

// ErrIPNotValid occurs when an Altar Broker is instantiated with an invalid IP address.
var ErrIPNotValid = errors.New("failed to initialise broker: IP address is not valid")

//nolint:tagliatelle
type awtrixConfig struct {
	// https://blueforcer.github.io/awtrix3/#/api?id=json-properties-1
	TimeAppEnabled bool `json:"TIM"`
}

// DisableAllDefaultApps configures the broker to diable all default apps on startup.
func DisableAllDefaultApps() func(*awtrixConfig) {
	return func(cfg *awtrixConfig) {
		defaultApps := []func(*awtrixConfig){
			DisableDefaultTimeApp(),
		}
		for _, fn := range defaultApps {
			fn(cfg)
		}
	}
}

// DisableDefaultTimeApp disables the default time app on the awtrix display on broker startup.
func DisableDefaultTimeApp() func(*awtrixConfig) {
	return func(cfg *awtrixConfig) {
		cfg.TimeAppEnabled = false
	}
}

// NewBroker instantiates a new Altar broker.
func NewBroker(
	addr string,
	applications []*application.Application,
	options ...func(*awtrixConfig),
) (*HTTPBroker, error) {
	if len(applications) == 0 {
		return nil, ErrBrokerHasNoApplications
	}

	clockIP := net.ParseIP(addr)
	if clockIP == nil {
		return nil, ErrIPNotValid
	}

	cfg := awtrixConfig{}
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
		slog.Error("error settin up initial awtrix configuration", "error", err)
	}

	go func() {
		for {
			startTime := time.Now()

			for _, app := range b.applications {
				err := app.Fetch()
				if err != nil {
					slog.Error("error fetching %v: %v", app.Name, err)
				}
			}

			for _, app := range b.applications {
				err := b.push(app)
				if err != nil {
					slog.Error("error pushing %v: %v", app.Name, err)
				}
			}

			duration := time.Since(startTime)
			if duration < MinLoopTime {
				time.Sleep(MinLoopTime - duration)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/shutdown", shutdownHandler)

	adminServer := &http.Server{
		Addr:         ":" + AdminPort,
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
		debugPort = ":8080"
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

	return err
}

func (b *HTTPBroker) push(app *application.Application) error {
	jsonData, err := json.Marshal(app.GetData())
	if err != nil {
		return fmt.Errorf("failed to marshal %v data into json: %w", app.Name, err)
	}

	bufferedJSON := bytes.NewBuffer(jsonData)

	debugPort := ""
	if b.Debug {
		debugPort = ":8080"
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
	// TODO: investigate why no error is printed when request has no response
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	return err
}

// TODO: have one general handler listen to four letter/worded commands.
func shutdownHandler(w http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)

		return
	}

	err = req.Body.Close()
	if err != nil {
		slog.Error("error in shutdown handler", "error", err)
	}

	if string(body) == "confirm" && req.Method == http.MethodPost {
		slog.Info("shutdown request received - shutting down")
		os.Exit(1)
	}
}
