// Package pipelinewatcher provides an example extension of altar's broker server.
package pipelinewatcher

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/t-monaghan/altar/application"
)

//nolint:gochecknoglobals
var (
	branchChannel      chan PullRequestActionsStatus
	once               sync.Once
	channelInitialized bool
)

func initChannel() {
	once.Do(func() {
		branchChannel = make(chan PullRequestActionsStatus, 5)
		channelInitialized = true
	})
}

// PipelineFetcher is the handler function for the Github pipeline application.
func PipelineFetcher(app *application.Application, client *http.Client) error {
	if !channelInitialized {
		initChannel()
	}

	var info PullRequestActionsStatus
	select {
	case info = <-branchChannel:
		slog.Debug("github pipeline fetcher received message from channel", "msg", info)
	default: // No data available in channel
		return nil
	}

	// write loading bar for in progress
	// write (completed)/(total) as text

	return nil
}

// PullRequestActionsStatus represents branch information from GitHub.
// TODO: restructure this as cli will pass pertinent info
type PullRequestActionsStatus struct {
	TotalActions     int      `json:"totalActions"`
	CompletedActions int      `json:"completedActions"`
	FailedActions    []string `json:"failedActions"`
}

// PipelineHandler handles HTTP requests with GitHub branch information.
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

	var branch PullRequestActionsStatus
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
