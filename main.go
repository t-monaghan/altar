// An example of Altar's intended usage
//
// To have this run in debug mode with a request logger run `devbox services up`
package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
)

func helloWorldFetcher(a *application.Application) error {
	a.Data.Text = "Hello, World!"

	return nil
}

func slowAppHandler(a *application.Application) error {
	a.Data.Text = a.PollRate.String()

	return nil
}

func main() {
	helloWorld := application.NewApplication("Hello World", helloWorldFetcher)
	slowApp := application.NewApplication("Slow App", slowAppHandler)
	slowApp.PollRate = time.Second * 30
	fastApp := application.NewApplication("Fast App", helloWorldFetcher)
	fastApp.PollRate = time.Second * 2
	appList := []*application.Application{&helloWorld, &slowApp, &fastApp}
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
