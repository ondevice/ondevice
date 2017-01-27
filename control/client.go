package control

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

func getJSON(tgt interface{}, endpoint string) error {
	transport := &http.Transport{
		Dial: func(proto, addr string) (conn net.Conn, err error) {
			return net.Dial(getSocketPath())
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
