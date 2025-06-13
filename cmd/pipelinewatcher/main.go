package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/t-monaghan/altar/broker"
	"github.com/t-monaghan/altar/examples/pipelinewatcher"
)

func main() {
	ownerRepoOut, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		slog.Error("failed to execute git command", "error", string(ownerRepoOut))
		os.Exit(1)
	}

	url := string(ownerRepoOut)
	ownerRepo := strings.Split(strings.Split(url, ":")[1], "/")
	branchOut, err := exec.Command("git", "branch", "--show-current").Output()
	branchName := strings.TrimSuffix(string(branchOut), "\n")

	if err != nil {
		slog.Error("failed to query current branch name", "error", branchName)
		os.Exit(1)
	}
	client := &http.Client{Timeout: time.Second * 5}
	branch := pipelinewatcher.GithubBranch{Owner: ownerRepo[0],
		Repo: strings.TrimSuffix(strings.TrimSuffix(ownerRepo[1], "\n"), ".git"), Branch: branchName}

	json, err := json.Marshal(branch)
	if err != nil {
		slog.Error("failed to marshal json body for github branch", "error", err)
		os.Exit(1)
	}

	body := bytes.NewBuffer(json)
	req, err := http.NewRequestWithContext(context.Background(),
		http.MethodPost, "http://127.0.0.1:"+broker.DefaultAdminPort+"/api/pipeline-watcher", body)

	if err != nil {
		slog.Error("failed to create request for github branch", "error", err)
		os.Exit(1)
	}

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("failed to perform request for github branch", "error", err)
		os.Exit(1)
	}

	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			slog.Error("failed to close response body", "error", err)
		}
	}()
}
