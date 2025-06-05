// An example of Altar's intended usage
//
// To have this run in debug mode with a request logger run `devbox services up`
package main

import (
	"log/slog"
	"os"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
)

func helloWorldFetcher() (string, error) {
	return "Hello, World!", nil
}

func main() {
	helloWorld := application.NewApplication("Hello World", helloWorldFetcher)
	appList := []*application.Application{&helloWorld}
	// TODO: read ip from config (viper?)
	// or allow dynamic address via HTTP request
	broker, err := broker.NewBroker(
		"127.0.0.1",
		appList,
		broker.DisableAllDefaultApps(),
	)
	// TODO: read debug from flag (cobra/viper?)
	broker.Debug = true

	if err != nil {
		slog.Error("error instantiating new broker", "error", err)
		os.Exit(1)
	}

	broker.Start()
}

// TODO: have a server spawn and listen to "/kill" to allow rolling a new broker build
