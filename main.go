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

func helloWorldFetcher(a *application.AppData) error {
	a.Text = "Hello, World!"

	return nil
}

func main() {
	helloWorld := application.NewApplication("Hello World", helloWorldFetcher)
	appList := []*application.Application{&helloWorld}
	broker, err := broker.NewBroker(
		"127.0.0.1",
		appList,
		broker.DisableAllDefaultApps(),
	)
	broker.Debug = true

	if err != nil {
		slog.Error("error instantiating new broker", "error", err)
		os.Exit(1)
	}

	broker.Start()
}
