package broker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/t-monaghan/altar/application"
)

type HttpBroker struct {
	applications []*application.Application
	clockAddress string
	client       *http.Client
	port         string
	Debug        bool
}

func NewBroker(clockIp net.IP, applications []*application.Application) (*HttpBroker, error) {
	if len(applications) == 0 {
		return nil, fmt.Errorf("failed to start broker: no applications were provided")
	}
	return &HttpBroker{
		clockAddress: fmt.Sprintf("http://%v", clockIp),
		applications: applications,
		client:       &http.Client{Timeout: 10 * time.Second},
		Debug:        false,
	}, nil
}

// TODO: how can this loop run whilst also providing endpoints?
// e.g. it would be useful to serve a /kill api for rollovers
// Have Start() in a goroutine, and spin the server up after
func (b *HttpBroker) Start() error {
	for {
		// Fetch
		for _, app := range b.applications {
			err := app.Fetch()
			if err != nil {
				fmt.Printf("error fetching %v: %v", app.Name, err)
			}
		}
		// Push
		for _, app := range b.applications {
			err := b.push(app)
			if err != nil {
				fmt.Printf("error pushing %v: %v", app.Name, err)
			}
		}
		// TODO: only sleep if time is less than a configured amount
		// as in, allow setting a min loop time
		time.Sleep(5 * time.Second)
	}

}

func (b *HttpBroker) push(app *application.Application) error {
	jsonData, err := json.Marshal(app.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal %v data into json: %w", app.Name, err)
	}
	bufferedJson := bytes.NewBuffer(jsonData)
	debugPort := ""
	if b.Debug {
		debugPort = ":8080"
	}
	address := fmt.Sprintf("%v%v/api/custom?name=%v", b.clockAddress, debugPort, url.QueryEscape(app.Name))
	req, err := http.NewRequest(http.MethodPost, address, bufferedJson)
	if err != nil {
		return fmt.Errorf("failed to create post request for %v: %w", app.Name, err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform post request for %v: %w", app.Name, err)
	}
	// TODO: investigate why no error is printed when request has no response
	return resp.Body.Close()
}
