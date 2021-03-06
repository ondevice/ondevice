package control

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ondevice/ondevice/config"
	"github.com/sirupsen/logrus"
)

// request -- control socket request
type request struct {
	Header   http.Header
	endpoint string

	body io.Reader
}

// Do -- run request request
func (r request) Do(method string) response {
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
	for strings.HasPrefix(r.endpoint, "/") {
		r.endpoint = r.endpoint[1:]
	}

	// TODO do proper URL parsing
	client := &http.Client{Transport: transport}

	var req, err = http.NewRequest(method, "http://ondevice/"+r.endpoint, r.body)
	if err != nil {
		return response{err: err}
	}

	if r.Header != nil {
		req.Header = r.Header
	}

	var resp *http.Response
	if resp, err = client.Do(req); err != nil {
		return response{resp: resp, err: err}
	}

	if resp.StatusCode != 200 {
		return response{
			resp: resp,
			err:  fmt.Errorf("Unexpected device request response code: %s", resp.Status),
		}
	}

	return response{resp: resp}
}

func (r request) Get() response {
	return r.Do("GET")
}

func (r request) PostForm(form url.Values) response {
	r.body = strings.NewReader(form.Encode())
	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r.Do("POST")
}

func getSocketURLs() []*url.URL {
	// deprecated - this will probably be removed (marked at 2020/08/05)
	if env := os.Getenv("ONDEVICE_HOST"); env != "" {
		// e.g.:
		// - unix:///var/run/ondevice/ondevice.sock
		// - /var/run/ondevice/ondevice.sock
		// - http://localhost:1234/

		u, err := url.Parse(env)
		if err != nil {
			logrus.WithError(err).Fatal("failed to parse ONDEVICE_HOST")
		}

		return []*url.URL{u}
	}

	var socketURL = config.MustLoad().GetPath(config.PathOndeviceSock)
	if socketURL.Error() != nil {
		logrus.WithError(socketURL.Error()).Fatal("failed to fetch daemon socket URL")
	}

	var rc []*url.URL
	for {
		if u := socketURL.GetAbsoluteURL(); u != nil {
			rc = append(rc, u)
		}
		if !socketURL.Next() {
			break
		}
	}

	return rc
}
