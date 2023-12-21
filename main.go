package main

import (
	"os"
	"github.com/jagheterfredrik/wallbox-mqtt-bridge/app"
)

func main() {
	if len(os.Args) != 2 {
		panic("Usage: ./bridge --config or ./bridge bridge.ini")
	}
	firstArgument := os.Args[1]
	if firstArgument == "--config" {
		bridge.RunTuiSetup()
		os.Exit(0)
	} else {
		bridge.LaunchBridge(firstArgument)
	}
}
