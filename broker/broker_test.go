package broker_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
	"github.com/t-monaghan/altar/utils"
	"github.com/t-monaghan/altar/utils/awtrix"
)

func Test_InvalidBrokerInstantiation(t *testing.T) {
	t.Parallel()

	toyApp := application.NewApplication("test app",
		func(_ *application.Application, _ *http.Client) error {
			return nil
		})
	toyAppList := []utils.AltarHandler{&toyApp}

	cases := []struct {
		description  string
		IPAddress    string
		Applications []utils.AltarHandler
		expected     error
	}{
		{"broker with no applications", "127.0.0.1", nil, broker.ErrBrokerHasNoApplications},
		{"broker with invalid IP", "foobarbaz", toyAppList, broker.ErrIPNotValid},
	}
	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			_, err := broker.NewBroker(testCase.IPAddress,
				testCase.Applications, map[string]func(http.ResponseWriter, *http.Request){})
			if err == nil || !errors.Is(err, testCase.expected) {
				t.Fatalf("did not throw expected error\n\texpected: %v\n\treceived: %v", testCase.expected, err)
			}
		})
	}
}

//nolint:gochecknoglobals
var empty200Response = &http.Response{
	StatusCode: http.StatusOK,
	Body:       io.NopCloser(bytes.NewBufferString("")),
}

func Test_BrokerHandlesRequests(t *testing.T) { //nolint:tparallel,funlen
	appMsg := "Hello, World!"
	appName := "test app"
	toyApp := application.NewApplication(appName,
		func(a *application.Application, _ *http.Client) error {
			a.Data.Text = appMsg

			return nil
		})
	toyAppList := []utils.AltarHandler{&toyApp}

	t.Run("broker executes handler requests", func(t *testing.T) {
		t.Parallel()

		brkr, err := broker.NewBroker("127.0.0.1", toyAppList, map[string]func(http.ResponseWriter, *http.Request){})
		if err != nil {
			t.Fatalf("should not throw error creating broker\n\treceived error: %v", err)
		}

		brkr.AdminPort = "54322"

		pushRequestCorrect := make(chan bool, 1)

		brkr.Client = utils.MockClient(func(request *http.Request) (*http.Response, error) {
			if request.URL.Path == "/api/settings" || request.URL.Path == "/api/reboot" {
				return empty200Response, nil
			}

			body, err := io.ReadAll(request.Body)
			if err != nil {
				t.Fatalf("failed reading body of broker request\n\terror: %v", err)
			}

			expected, _ := json.Marshal(application.AppData{Text: appMsg})
			if string(body) != string(expected) {
				t.Fatalf("broker sent request with incorrect body\n\texpected: %v\n\treceived: %v", string(expected), string(body))
			}

			queries := request.URL.Query()
			if queries.Get("name") != appName {
				t.Fatalf("incorrect query paramater for app name\n\texpected: %v\n\treceived: %v", appName, queries["name"])
			}

			pushRequestCorrect <- true

			return empty200Response, nil
		})

		_, cancel := context.WithCancel(t.Context())
		go func() {
			defer cancel()
			brkr.Start()
		}()

		select {
		case rcv := <-pushRequestCorrect:
			if rcv != true {
				t.Fatal("push request is incorrect")
			}
		case <-time.After(time.Second * 3):
			t.Fatal("timed out waiting for broker to push message")
		}

		cancel()
		shutdownBroker(t, brkr)
	})
}

//nolint:funlen
func Test_BrokerSetsConfig(t *testing.T) {
	t.Parallel()

	appMsg := "Hello, World!"
	appName := "test app"
	toyApp := application.NewApplication(appName,
		func(a *application.Application, _ *http.Client) error {
			a.Data.Text = appMsg

			return nil
		})
	toyAppList := []utils.AltarHandler{&toyApp}

	cases := []struct {
		description string
		configFn    func() func(*awtrix.Config)
		expected    string
		port        string
	}{
		{"some description", application.DisableDefaultTimeApp, "{\"TIM\":false}", "43324"},
		{"disable all default apps", application.DisableAllDefaultApps,
			"{\"TIM\":false,\"WD\":false,\"DAT\":false,\"HUM\":false,\"TEMP\":false,\"BAT\":false}", "43325"},
	}
	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			configRequestReceived := make(chan struct{})
			configRequestCorrect := make(chan bool, 1)

			brkr, err := broker.NewBroker("127.0.0.1", toyAppList,
				map[string]func(http.ResponseWriter, *http.Request){}, testCase.configFn())
			if err != nil {
				t.Fatalf("should not throw error creating broker\n\treceived error: %v", err)
			}

			brkr.AdminPort = testCase.port

			var settingsRequestCount int32

			brkr.Client = utils.MockClient(func(request *http.Request) (*http.Response, error) {
				// we only care about the first settings request (initial config)
				if request.URL.Path == "/api/settings" && atomic.AddInt32(&settingsRequestCount, 1) == 1 {
					body, err := io.ReadAll(request.Body)
					if err != nil {
						t.Fatalf("failed reading body of broker request\n\terror: %v", err)
					}

					configRequestCorrect <- string(body) == testCase.expected

					close(configRequestReceived)

					if string(body) != testCase.expected {
						t.Logf(
							"broker sent request with incorrect body\n\texpected: %v\n\treceived: %v",
							testCase.expected,
							string(body))
					}
				}

				return empty200Response, nil
			})

			_, cancel := context.WithCancel(t.Context())
			go func() {
				defer cancel()
				brkr.Start()
			}()

			select {
			case <-configRequestReceived:
				result := <-configRequestCorrect
				if !result {
					t.Fatalf("broker sent incorrect config request body")
				}
			case <-time.After(3 * time.Second):
				t.Fatalf("timed out waiting for config request")
			}

			cancel()
			shutdownBroker(t, brkr)
		})
	}
}

func shutdownBroker(t *testing.T, brkr *broker.HTTPBroker) {
	t.Helper()

	realClient := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(
		t.Context(),
		http.MethodPost,
		"http://localhost:"+brkr.AdminPort,
		bytes.NewBufferString(string(broker.AdminShutdownCommand)),
	)

	if err != nil {
		t.Fatalf("should not throw error creating shutdown request\n\treceived error: %v", err)
	}

	resp, err := realClient.Do(req)
	if err != nil {
		t.Fatalf("failed to send shutdown request\n\terror: %v", err)
	}

	func() {
		if err := resp.Body.Close(); err != nil {
			t.Fatalf("error closing response body: %v", err)
		}
	}()
}
