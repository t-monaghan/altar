package main

import (
	"fmt"
	"net"
	"os"

	"github.com/t-monaghan/altar/application"
	"github.com/t-monaghan/altar/broker"
)

func helloWorldFetcher() (string, error) {
	return "Hello, World!", nil
}

var HelloWorld = application.NewApplication("Hello World", helloWorldFetcher)

func main() {
	appList := []*application.Application{&HelloWorld}
	// TODO: read ip from config (viper?)
	broker, err := broker.NewBroker(net.ParseIP("127.0.0.1"), appList)
	// TODO: read debug from flag (cobra/viper?)
	broker.Debug = true
	if err != nil {
		fmt.Printf("error instantiating new broker: %v", err)
		os.Exit(1)
	}
	err = broker.Start()
	if err != nil {
		fmt.Printf("broker encountered an error during runtime: %v", err)
		os.Exit(1)
	}
	// TODO: have a server spawn and listen to "/kill" to allow rolling a new broker build
}
