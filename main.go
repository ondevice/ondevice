/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import "log"
import "net/http"
import "time"
import "github.com/ondevice/ondevice/cmd"
import "github.com/ondevice/ondevice/config"

func main() {
	// disable date/time logging (there's an override for `ondevice daemon`)
	log.SetFlags(0)

	// set a default timeout of 30sec for REST API calls (will be reset in long-running commands)
	// TODO use a builder pattern to be able to specify this on a per-request basis
	// Note: doesn't affect websocket connections
	var timeout = time.Duration(config.GetInt("client", "timeout", 30))
	http.DefaultClient.Timeout = timeout * time.Second
	cmd.Execute()
}
