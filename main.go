// An example of Altar's intended usage
//
// To have this run in debug mode with a request logger run `devbox services up`
package main

import (
	"fmt"
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

func randomlySkip(a *application.Application) error {
	r := time.Now().Second()
	if r%2 != 0 {
		a.PushOnNextCall = false
	}

	return nil
}

func throwsErrors(_ *application.Application) error {
	return fmt.Errorf("an example of an error")
}

func main() {
	slowApp := application.NewApplication("Slow App", helloWorldFetcher)
	slowApp.PollRate = time.Second * 30

	fastApp := application.NewApplication("Fast App", helloWorldFetcher)
	fastApp.PollRate = time.Second * 2

	inconsistentApp := application.NewApplication("Inconsistent App", randomlySkip)
	inconsistentApp.PollRate = time.Second * 5

	erroringApp := application.NewApplication("Throws Errors", throwsErrors)

	appList := []*application.Application{&slowApp, &fastApp, &inconsistentApp, &erroringApp}
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
