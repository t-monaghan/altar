// Package pipelinewatcher provides an example extension of altar's broker server.
package pipelinewatcher

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/t-monaghan/altar/application"
)

var branchChannel = make(chan GithubBranch, 5)

func PipelineFetcher(app *application.Application, client *http.Client) error {
	// receive information from channel
	// write loading bar for in progress
	// write (completed)/(total) as text
	select {
	case branch := <-branchChannel:
		slog.Debug("github pipeline fetcher received message from channel", "msg", branch)
		app.Data.Text = branch.Repo
	default:

	}
	// seems push on next call should be decided by whether

	return nil
}

// func NotifyFailedRuns(ntfr *notifier.Notifier, client *http.Client) error {
// // post failed runs as notification?
// }

// func NotifyNewCommit/Run(ntfr *notifier.Notifier, client *http.Client) error {
// // post failed runs as notification?
// }

type GithubBranch struct {
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
}

func PipelineHandler(rsp http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		slog.Error("github pipeline listener failed to read body", "error", err)
	}
	slog.Debug("github pipeline listener received message", "body", string(body))

	var branch GithubBranch
	err = json.Unmarshal(body, &branch)

	if err != nil {
		slog.Error("github pipeline listener failed to unmarshal request", "error", err)
	}

	slog.Debug("github pipeline listener posted message to channel", "msg", branch)
	branchChannel <- branch
	// this request is created by cli tool?
	// pipe into channel details for fetcher
	rsp.WriteHeader(http.StatusOK)
}
