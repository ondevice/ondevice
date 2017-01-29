package control

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

func getSocketURLs() []url.URL {
	if env := os.Getenv("ONDEVICE_HOST"); env != "" {
		// e.g.:
		// - unix:///var/run/ondevice/ondevice.sock
		// - /var/run/ondevice/ondevice.sock
		// - http://localhost:1234/

		u, err := url.Parse(env)
		if err != nil {
			logg.Fatal("Failed to parse ONDEVICE_HOST: ", err)
		}

		return []url.URL{*u}
	}

	return []url.URL{
		url.URL{Scheme: "unix", Path: config.GetConfigPath("ondevice.sock")},
		url.URL{Scheme: "unix", Path: "/var/run/ondevice/ondevice.sock"},
	}
}

func getJSON(tgt interface{}, endpoint string) error {
	transport := &http.Transport{
		Dial: func(proto, addr string) (conn net.Conn, err error) {
			urls := getSocketURLs()
			var firstError error

			for _, url := range urls {
				var protocol, path string

				if url.Scheme == "unix" || url.Scheme == "" {
					protocol = "unix"
					path = url.Path
				} else if url.Scheme == "http" {
					protocol = "tcp"
					path = url.Host
				}

				c, err := net.Dial(protocol, path)
				if err == nil {
					// it worked
					return c, nil
				} else if firstError == nil {
					firstError = err
				}
			}

			return nil, firstError
		},
	}

	// remove leading slashes
	for strings.HasPrefix(endpoint, "/") {
		endpoint = endpoint[1:]
	}

	// TODO do proper URL parsing
	client := &http.Client{Transport: transport}
	resp, err := client.Get("http://ondevice/" + endpoint)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Unexpected device request response code: %s", resp.Status)
	}

	buff := make([]byte, resp.ContentLength)
	count, err := resp.Body.Read(buff)
	if err == io.EOF && int64(count) == resp.ContentLength {
		// ignored
	} else if err != nil {
		return err
	}

	json.Unmarshal(buff[:count], &tgt)
	return nil
}

// GetState -- Query the device state over the unix socket
func GetState() (DeviceState, error) {
	rc := DeviceState{}
	err := getJSON(&rc, "/state")
	return rc, err
}
