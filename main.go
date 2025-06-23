// An example of altar's intended usage
//
// To have this run in debug mode with a request logger run `devbox services up`
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
	"github.com/t-monaghan/altar/examples/buttons"
	"github.com/t-monaghan/altar/examples/github/checks"
	"github.com/t-monaghan/altar/examples/github/contributions"
	"github.com/t-monaghan/altar/examples/weather"
	"github.com/t-monaghan/altar/notifier"
	"github.com/t-monaghan/altar/utils"
)

func main() {
	githubChecks := notifier.NewNotifier("github checks", checks.Fetcher)
	weather := application.NewApplication("rain forecast", weather.Fetcher)
	githubContributions := application.NewApplication("github contributions", contributions.Fetcher)

	handlers := map[string]func(http.ResponseWriter, *http.Request){
		"/api/pipeline-watcher": checks.Handler,
		"/api/contributions":    contributions.Handler,
		"/api/buttons":          buttons.Handler,
	}

	appList := []utils.Routine{&githubChecks, &weather, &githubContributions}

	checkRequiredEnvironmentVariables()

	brkr, err := broker.NewBroker(
		"127.0.0.1",
		appList,
		handlers,
		broker.DisableAllDefaultApps(),
	)

	brkr.DebugMode = true

	if err != nil {
		slog.Error("error instantiating new broker", "error", err)
		os.Exit(1)
	}

	brkr.Start()
}

func checkRequiredEnvironmentVariables() {
	requiredEnvVars := []string{"LATITUDE", "LONGITUDE"}
	missingVars := []string{}

	for _, val := range requiredEnvVars {
		if os.Getenv(val) == "" {
			missingVars = append(missingVars, val)
		}
	}

	if len(missingVars) > 0 {
		slog.Error("missing required environment variables", "missing-env-vars", missingVars)
		os.Exit(1)
	}
}
