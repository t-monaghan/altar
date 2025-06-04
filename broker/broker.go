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

const minLoopTime = 5 * time.Second
const httpTimeout = 10 * time.Second
const idleTimeout = 120 * time.Second
const adminPort = "25827"

// HTTPBroker is a broker that queries each of the Altar applications and communicates updates to the Awtrix host.
type HTTPBroker struct {
	applications []*application.Application
	clockAddress string
	client       *http.Client
	Debug        bool
}

// ErrBrokerHasNoApplicationsError occurs when an Altar Broker is instantiated with no applications.
var ErrBrokerHasNoApplicationsError = errors.New("failed to start broker: no applications were provided")

// NewBroker instantiates a new Altar broker.
func NewBroker(clockIP net.IP, applications []*application.Application) (*HTTPBroker, error) {
	if len(applications) == 0 {
		return nil, ErrBrokerHasNoApplicationsError
	}

	return &HTTPBroker{
		clockAddress: fmt.Sprintf("http://%v", clockIP),
		applications: applications,
		client:       &http.Client{Timeout: httpTimeout},
		Debug:        false,
	}, nil
}

// Start executes the broker's routine.
func (b *HTTPBroker) Start() {
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
			if duration < minLoopTime {
				time.Sleep(minLoopTime - duration)
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/shutdown", shutdownHandler)

	adminServer := &http.Server{
		Addr:         ":" + adminPort,
		Handler:      mux,
		ReadTimeout:  httpTimeout,
		WriteTimeout: httpTimeout,
		IdleTimeout:  idleTimeout,
	}

	log.Fatal(adminServer.ListenAndServe())
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

	resp, err := b.client.Do(req)
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

	if string(body) == "confirm" {
		slog.Info("shutdown request received - shutting down")
		os.Exit(1)
	}
}
