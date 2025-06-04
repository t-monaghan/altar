package broker_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
	"github.com/t-monaghan/altar/utils"
)

func Test_InvalidBrokerInstantiation(t *testing.T) {
	t.Parallel()

	toyApp := application.NewApplication("test app", func() (string, error) { return "", nil })
	toyAppList := []*application.Application{&toyApp}

	cases := []struct {
		description  string
		IPAddress    string
		Applications []*application.Application
		expected     error
	}{
		{"broker with no applications", "127.0.0.1", nil, broker.ErrBrokerHasNoApplications},
		{"broker with invalid IP", "foobarbaz", toyAppList, broker.ErrIPNotValid},
	}
	for _, testCase := range cases {
		t.Run(testCase.description, func(t *testing.T) {
			t.Parallel()

			_, err := broker.NewBroker(testCase.IPAddress, testCase.Applications)

			if err == nil || !errors.Is(err, testCase.expected) {
				t.Fatalf("did not throw expected error\n\texpected: %v\n\treceived: %v", testCase.expected, err)
			}
		})
	}
}

func Test_BrokerHandlesRequests(t *testing.T) { //nolint:tparallel
	appMsg := "Hello, World!"
	appName := "test app"
	toyApp := application.NewApplication(appName, func() (string, error) { return appMsg, nil })
	toyAppList := []*application.Application{&toyApp}

	t.Run("broker executes handler requests", func(t *testing.T) {
		t.Parallel()

		testBroker, err := broker.NewBroker("127.0.0.1", toyAppList)
		if err != nil {
			t.Fatalf("should not throw error creating broker\n\treceived error: %v", err)
		}

		testBroker.Client = utils.MockClient(func(request *http.Request) (*http.Response, error) {
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

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString("")),
			}, nil
		})
		go func() {
			testBroker.Start()
		}()

		// TODO: how can I have a successful request propagate a "pass" message from the goroutine?
		// OR is the goroutine the wrong pattern here?
		time.Sleep(time.Second)

		realClient := &http.Client{Timeout: 10 * time.Second}
		req, err := http.NewRequestWithContext(
			t.Context(),
			http.MethodPost,
			"http://localhost:"+broker.AdminPort,
			bytes.NewBufferString("confirm"),
		)

		if err != nil {
			t.Fatalf("should not throw error creating shutdown request\n\treceived error: %v", err)
		}

		resp, err := realClient.Do(req)
		if err != nil {
			t.Fatalf("failed to send shutdown request\n\terror: %v", err)
		}
		defer resp.Body.Close() //nolint:errcheck
	})
}
