// Package pipelinewatcher provides an example extension of altar's broker server.
package pipelinewatcher

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
	branchChannel      chan ChecksProgress
	once               sync.Once
	channelInitialized bool
)

const channelBufferSize = 5

func initChannel() {
	once.Do(func() {
		branchChannel = make(chan ChecksProgress, channelBufferSize)
		channelInitialized = true
	})
}

// PipelineFetcher is the handler function for the Github pipeline application.
func PipelineFetcher(ntfr *notifier.Notifier, _ *http.Client) error {
	if !channelInitialized {
		initChannel()
	}

	var info ChecksProgress
	select {
	case info = <-branchChannel:
		slog.Debug("github pipeline fetcher received message from channel", "msg", info)
	default: // No data available in channel
		ntfr.PushOnNextCall = false

		return nil
	}

	progress := info.CompletedActions / info.TotalActions
	ntfr.Data.Progress = &progress
	ntfr.Data.ProgressC = []int{74, 194, 108}
	ntfr.Data.ProgressBC = []int{17, 99, 42}

	if len(info.FailedActions) > 0 {
		fiveHundred := 500
		trueVal := true
		ntfr.Data.BlinkText = &fiveHundred
		ntfr.Data.Hold = &trueVal
		ntfr.Data.Text = fmt.Sprintf("%v failed", info.FailedActions[0])

		if len(info.FailedActions) > 1 {
			ntfr.Data.Text = fmt.Sprintf("%v failing", len(info.FailedActions))
		}

		return nil
	}

	ntfr.Data.Text = fmt.Sprintf("%v/%v jobs", info.CompletedActions, info.TotalActions)

	return nil
}

// ChecksProgress represents the progress of required checks for a given PR.
type ChecksProgress struct {
	TotalActions     int      `json:"totalActions"`
	CompletedActions int      `json:"completedActions"`
	FailedActions    []string `json:"failedActions"`
}

// PipelineHandler handles HTTP requests with GitHub checks information.
func PipelineHandler(rsp http.ResponseWriter, req *http.Request) {
	// Ensure channel is initialized
	if !channelInitialized {
		initChannel()
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("github pipeline listener failed to read body", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	slog.Debug("github pipeline listener received message", "body", string(body))

	var branch ChecksProgress
	err = json.Unmarshal(body, &branch)

	if err != nil {
		slog.Error("github pipeline listener failed to unmarshal request", "error", err)
		rsp.WriteHeader(http.StatusBadRequest)

		return
	}

	slog.Debug("github pipeline listener posted message to channel", "msg", branch)

	// Use non-blocking send to avoid deadlocks if channel is full
	select {
	case branchChannel <- branch:
		// Successfully sent
	default:
		slog.Warn("github pipeline channel is full, dropping message")
	}

	rsp.WriteHeader(http.StatusOK)
}

// Reset clears the channel state - useful for testing.
func Reset() {
	if channelInitialized {
		// Clear the channel
		for len(branchChannel) > 0 {
			<-branchChannel
		}
	}
}
