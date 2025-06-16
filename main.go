// An example of Altar's intended usage
//
// To have this run in debug mode with a request logger run `devbox services up`
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
	"github.com/t-monaghan/altar/examples/pipelinewatcher"
	"github.com/t-monaghan/altar/examples/weather"
	"github.com/t-monaghan/altar/notifier"
	"github.com/t-monaghan/altar/utils"
)

func main() {
	githubApp := notifier.NewNotifier("github", pipelinewatcher.PipelineFetcher)
	weatherApp := application.NewApplication("Rain Forecast", weather.RainChanceFetcher)

	appList := []utils.AltarHandler{&weatherApp, &githubApp}

	listeners := map[string]func(http.ResponseWriter, *http.Request){
		"/api/pipeline-watcher": pipelinewatcher.PipelineHandler}

	broker, err := broker.NewBroker(
		"127.0.0.1",
		appList,
		listeners,
		application.DisableAllDefaultApps(),
	)

	broker.Debug = true

	if err != nil {
		slog.Error("error instantiating new broker", "error", err)
		os.Exit(1)
	}

	broker.Start()
}
