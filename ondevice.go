package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ondevice/ondevice/command"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

func main() {
	// disable date/time logging (there's an override for `ondevice daemon`)
	log.SetFlags(0)

	if len(os.Args) < 2 {
		logg.Fatal("Missing command! try 'ondevice help'")
	}

	// set a default timeout of 30sec for REST API calls (will be reset in long-running commands)
	// TODO use a builder pattern to be able to specify this on a per-request basis
	// Note: doesn't affect websocket connections
	var timeout = time.Duration(config.GetInt("client", "timeout", 30))
	http.DefaultClient.Timeout = timeout * time.Second

	//logg.Debug("-- args: ", os.Args[1:])
	cmd := os.Args[1]
	rc := command.Run(cmd, os.Args[2:])
	os.Exit(rc)
}
