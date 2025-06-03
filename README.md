# **A**wtrix **L**istens **T**o **A**ltar **R**equests

> [!WARNING]
> This project is still a work in progress and is not yet ready for use.

## Summary

Altar is a framework that allows developers to create custom Applications for the [Awtrix](https://blueforcer.github.io/awtrix3/#/) platform. It's aim is to provide a simple and intuitive manner to stand up a broker that can fetch data from various sources and display it on Awtrix supported devices.

### See it in action

You don't need an awtrix device to see this run. To have a look at what's going on you just need [devbox](https://www.jetify.com/devbox/). After installing devbox simply run `devbox services up` in this project and you can see the requests from the example application hitting the request-debugger process.

## Using the library

### Getting Started

First define an Application:

```go filename="hello.go"
// for brevity we're defining the application in main
// defining applications in a separate package is recommended
package main

import "github.com/t-monaghan/altar/application"

func helloWorldFetcher() (string, error) {
	return "Hello, World!", nil
}

var HelloWorld = application.NewApplication("Hello World", helloWorldFetcher)
```

Then define the main function, starting the broker with a list including this application:

```go filename="main.go"
package main

import (
	"fmt"
	"net"
	"os"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
)

func main() {
	appList := []*application.Application{HelloWorld}
	broker, err := broker.NewBroker(net.ParseIP("YOUR_AWTRIX_IP_HERE"), appList)
	if err != nil {
		fmt.Printf("error instantiating new broker: %v", err)
		os.Exit(1)
	}
	err = broker.Start()
	if err != nil {
		fmt.Printf("broker encountered an error during runtime: %v", err)
		os.Exit(1)
	}
}
```

Finally, build and run the broker:

```sh
go build
./altar
```

The broker will handle pulling down and standing up the new custom applications, however new applications require manually shutting the broker down and starting the new build. I am planning to develop a launcher that will handle the rollover of new broker builds.

## Contributing

This project is not currently accepting contributions, however I will be streaming my development of this project through a Zed channel [here](https://zed.dev/channel/altar-22876).
