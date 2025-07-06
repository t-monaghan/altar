![Curved altar from the video game Runescape](img/banner.png)

# **A**wtrix **L**istens **T**o **A**ltar **R**equests

## Summary

Altar is a framework to control [Awtrix](https://blueforcer.github.io/awtrix3/#/) displays.

### How?

Awtrix devices are thin clients, exposing an API that will let you manipulate the data of an Application or Notification. Altar builds on this providing a way to communicate between the outside world and the awtrix device. Altar does this by providing a broker, which has a set of Application and Notification routines, as well as a set of http handlers. Routines can be as simple as [polling the weather](examples/weather) or can be more involved such as being [triggered by a cli to watch GitHub action runs](examples/github/checks).

### See it in action

You don't need an awtrix device to run this project. To see it in action you can install [devbox](https://www.jetify.com/devbox/) and run `devbox services up --process-compose-file scripts/request-logger-pc.yaml` to see the requests from the example application get captured by a request logger.

> [!WARNING]
> Devbox won't start a new shell without a file at `.env`, this is an issue I've raised with devbox [here](https://github.com/issues/created?issue=jetify-com%7Cdevbox%7C2504). You can run `cp .env.example .env` to have this file created with some defaults.

## Using the library

### Quickstart

First define an Application:

```go
package main

import "github.com/t-monaghan/altar/application"

func helloWorldFetcher(app *application.Application, _ *http.Client) error {
	app.Data.Text = "Hello, World!"
	return nil
}

var HelloWorld = application.NewApplication("Hello World", helloWorldFetcher)
```

Then define the main function, starting the broker with a list containing the above application:

```go
package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
)

func main() {
	helloWorld := application.NewApplication("Hello World", helloWorldFetcher)
	appList := []*application.Application{&helloWorld}
	broker, err := broker.NewBroker(
		"YOUR_AWTRIX_IP_HERE",
		appList,
		nil,
	)

	if err != nil {
		slog.Error("error instantiating new broker", "error", err)
		os.Exit(1)
	}

	broker.Start()
}
```

Finally, run your program:

```sh
go run .
```

### Going deeper

Routines with more functionality can be found in the [examples](https://github.com/t-monaghan/altar/tree/main/examples) package.

## Running locally

This project requires some environment variables to be set for it to be run locally, there is an example dotenv file with some defaults to get you started quickly. To use this example you can run `cp .env.example .env`. The required environment variables are explained within this example file.

### Common issues

If you get the error `Error: failed parsing .env file. Error: failed to open file: .../altar/.env` please create a `.env` file with the required environment variables. This can be done with `cp .env.example .env`, then to enter the devbox shell you can run `direnv reload` or `devbox shell`.

## Contributing

This project is not currently accepting contributions, however I will be streaming my development of this project through a Zed channel [here](https://zed.dev/channel/altar-22876).
