# **A**wtrix **L**istens **T**o **A**ltar **R**equests

## Summary

Altar is a framework that allows developers to create custom Applications for the Awtrix platform. It's aim is to provide a simple and intuitive manner to stand up a server that can fetch data from various sources and display it on Awtrix supported devices.

## Getting Started

First define an Application:

```go
// for brevity we're defining the application in main
// defining applications in a seperate package is recommended
package main

import (
	"github.com/t-monaghan/altar"
)

func helloWorldFetcher() (string, error) {
	return "Hello, World!", nil
}

var HelloWorld = altar.Application{
	Name:        "Hello World",
	Fetcher:     helloWorldFetcher,
}
```

Then define the main function, starting the server with a list including this application:

```go
package main

import (
	"github.com/t-monaghan/altar"
)

func main() {
	appList := []altar.Application{HelloWorld}
	server := altar.NewServer()
	server.Start(appList)
}
```

Finally, build and run the server:

```sh
go build
./altar
```

The server will handle pulling down and standing up the new custom applications, however I am planning to develop a launcher that will handle the rollover of new server builds.

## Contributing

This project is not currently accepting contributions, however I will be streaming my development of this project through a Zed channel [here](https://zed.dev/channel/altar-22876).
