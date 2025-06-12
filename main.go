// An example of Altar's intended usage
//
// To have this run in debug mode with a request logger run `devbox services up`
package main

import (
	"log/slog"
	"os"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
	"github.com/t-monaghan/altar/examples/weather"
	"github.com/t-monaghan/altar/utils"
)

func main() {
	weatherApp := application.NewApplication("Rain Forecast", weather.RainChanceFetcher)

	appList := []utils.AltarHandler{&weatherApp}

	broker, err := broker.NewBroker(
		"127.0.0.1",
		appList,
		application.DisableAllDefaultApps(),
	)

	broker.Debug = true

	if err != nil {
		slog.Error("error instantiating new broker", "error", err)
		os.Exit(1)
	}

	broker.Start()
}
