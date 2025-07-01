// Package stars provides an example extension of altar's broker server.
package stars

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/t-monaghan/altar/notifier"
)

//nolint:gochecknoglobals
var (
	starrerChannel     chan string
	once               sync.Once
	channelInitialized bool
)

const channelBufferSize = 5

func initChannel() {
	once.Do(func() {
		starrerChannel = make(chan string, channelBufferSize)
		channelInitialized = true
	})
}

// Fetcher receives data from the handler and prepares it to be posted by altar's broker.
func Fetcher(ntfr *notifier.Notifier, _ *http.Client) error {
	if !channelInitialized {
		initChannel()
	}
	var starrer string
	select {
	case starrer = <-starrerChannel:
		slog.Debug("contributions fetcher received contributions count", "starrer", starrer)
	default:
		ntfr.PushOnNextCall = false

		return nil
	}
	ntfr.PushOnNextCall = true
	ntfr.Data.Text = starrer

	return nil
}

type GitHubWebhook struct {
	Payload string `json:"payload"`
}

type WebhookPayload struct {
	Sender Sender `json:"sender"`
	Action string `json:"action"`
}

type Sender struct {
	Login string `json:"login"`
}

// Handler receives data from the gh-altar tool and passes it onto Fetcher.
func Handler(rsp http.ResponseWriter, req *http.Request) {
	if !channelInitialized {
		initChannel()
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("github checks handler failed to read body", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	// First, parse the outer JSON to get the payload string
	var githubWebhook GitHubWebhook
	if err := json.Unmarshal(body, &githubWebhook); err != nil {
		slog.Error("failed to unmarshal outer webhook JSON", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)
		return
	}

	// Then parse the payload string to get the actual data
	var webhookPayload WebhookPayload
	if err := json.Unmarshal([]byte(githubWebhook.Payload), &webhookPayload); err != nil {
		slog.Error("failed to unmarshal payload JSON", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)
		return
	}

	if webhookPayload.Action != "created" {
		slog.Warn("action", "action", webhookPayload.Action)
		return
	}

	select {
	case starrerChannel <- webhookPayload.Sender.Login:
	default:
		slog.Warn("github checks channel is full, dropping message")
	}

	rsp.WriteHeader(http.StatusOK)
}

// Reset clears the state of the channel used to communicate between the api handler and the altar fetcher.
func Reset() {
	if channelInitialized {
		for len(starrerChannel) > 0 {
			<-starrerChannel
		}
	}
}
