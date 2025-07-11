// Package checks provides an example extension of altar's broker server.
package checks

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/t-monaghan/altar/notifier"
)

//nolint:gochecknoglobals
var (
	checksChannel      chan Progress
	once               sync.Once
	channelInitialized bool
)

const channelBufferSize = 5

func initChannel() {
	once.Do(func() {
		checksChannel = make(chan Progress, channelBufferSize)
		channelInitialized = true
	})
}

// Fetcher receives data from the handler and prepares it to be posted by altar's broker.
func Fetcher(ntfr *notifier.Notifier, _ *http.Client) error {
	falseVal := false
	trueVal := true

	if !channelInitialized {
		initChannel()
	}

	progressOutOfAHundred := 0

	var info Progress

	select {
	case info = <-checksChannel:
		slog.Debug("githubchecks fetcher received message", "msg", info)
	default:
		ntfr.PushOnNextCall = false

		return nil
	}

	progressOutOfAHundred = int(float64(info.CompletedActions) / float64(info.TotalActions) * 100) //nolint:mnd

	if ntfr.Data.Progress != nil && progressOutOfAHundred == *ntfr.Data.Progress {
		return nil
	}

	ntfr.Data.Progress = &progressOutOfAHundred
	ntfr.PushOnNextCall = true
	ntfr.Data.Stack = &falseVal

	ntfr.Data.ProgressC = []int{74, 194, 108}
	ntfr.Data.ProgressBC = []int{17, 99, 42}
	eight := 8
	ntfr.Data.Duration = &eight

	if len(info.FailedActions) > 0 {
		fiveHundred := 800
		ntfr.Data.BlinkText = &fiveHundred
		ntfr.Data.Color = []int{255, 0, 0}
		ntfr.Data.Hold = &trueVal
		ntfr.Data.Text = fmt.Sprintf("%v failed", info.FailedActions[0])

		if len(info.FailedActions) > 1 {
			ntfr.Data.Text = fmt.Sprintf("%v failing", len(info.FailedActions))
		}

		return nil
	}

	if progressOutOfAHundred == 100 { //nolint:mnd
		ntfr.Data.Hold = &trueVal
		ntfr.Data.Text = "passing"
		ntfr.Data.Color = []int{0, 190, 0}

		return nil
	}

	ntfr.Data.Text = fmt.Sprintf("%v/%v jobs", info.CompletedActions, info.TotalActions)

	return nil
}

// Progress represents the progress of required checks for a given PR.
type Progress struct {
	TotalActions     int      `json:"totalActions"`
	CompletedActions int      `json:"completedActions"`
	FailedActions    []string `json:"failedActions"`
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

	slog.Debug("github checks handler received message", "body", string(body))

	var checks Progress
	err = json.Unmarshal(body, &checks)

	if err != nil {
		slog.Error("github checks handler failed to unmarshal request", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	slog.Debug("github checks handler posted message to channel", "msg", checks)

	select {
	case checksChannel <- checks:
	default:
		slog.Warn("github checks channel is full, dropping message")
	}

	rsp.WriteHeader(http.StatusOK)
}

// Reset clears the state of the channel used to communicate between the api handler and the altar fetcher.
func Reset() {
	if channelInitialized {
		for len(checksChannel) > 0 {
			<-checksChannel
		}
	}
}
