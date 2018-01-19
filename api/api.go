package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

type ErrorMessage struct {
	Status string `json:"status"`
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
}

func delete(endpoint string, params map[string]string, bodyType string, body []byte, auths ...Authentication) (*http.Response, error) {
	return _request("DELETE", endpoint, params, bodyType, body, auths...)
}

func get(endpoint string, params map[string]string, auths ...Authentication) (*http.Response, error) {
	return _request("GET", endpoint, params, "", nil, auths...)
}

func post(endpoint string, params map[string]string, bodyType string, body []byte, auths ...Authentication) (*http.Response, error) {
	return _request("POST", endpoint, params, bodyType, body, auths...)
}

func _request(method string, endpoint string, params map[string]string, bodyType string, body []byte, auths ...Authentication) (*http.Response, error) {
	var auth *Authentication

	if auths == nil {
		a, _ := CreateClientAuth()
		auth = &a
	} else {
		auth = &auths[0]
	}

	url := auth.GetURL(endpoint, params, "https")
	logg.Debugf("%s request to URL %s\n", method, url)
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		// TODO return err
		logg.Fatal("Failed to create request", err)
	}
	req.Header.Add("Authorization", auth.GetAuthHeader())
	req.Header.Add("User-agent", fmt.Sprintf("ondevice v%s", config.GetVersion()))

	if body != nil {
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		// TODO make the content type configurable
		req.Header.Add("Content-type", bodyType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, err
}

func deleteBody(endpoint string, params map[string]string, bodyType string, body []byte, auths ...Authentication) ([]byte, error) {
	return _getBody(delete(endpoint, params, bodyType, body, auths...))
}

func getBody(endpoint string, params map[string]string, auths ...Authentication) ([]byte, error) {
	return _getBody(get(endpoint, params, auths...))
}

func postBody(endpoint string, params map[string]string, bodyType string, body []byte, auths ...Authentication) ([]byte, error) {
	return _getBody(post(endpoint, params, bodyType, body, auths...))
}

func _getBody(resp *http.Response, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}

	var errMsg ErrorMessage
	if resp.StatusCode != http.StatusOK {
		errMsg = getErrorMessage(resp)

		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("Authentication failed: %s", errMsg.Msg)
		} else if resp.StatusCode == http.StatusTooManyRequests {
			var delayStr = resp.Header.Get("X-Ratelimit-Delay")
			return nil, fmt.Errorf("Error: Too many requests (try again in %ss)", delayStr)
		}

		// else
		return nil, fmt.Errorf("Request error (code %d): %s", errMsg.Code, errMsg.Msg)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func deleteObject(tgtValue interface{}, endpoint string, params map[string]string, bodyType string, body []byte, auths ...Authentication) error {
	body, err := deleteBody(endpoint, params, bodyType, body, auths...)
	return _getObject(tgtValue, body, err)
}

func getObject(tgtValue interface{}, endpoint string, params map[string]string, auths ...Authentication) error {
	body, err := getBody(endpoint, params, auths...)
	return _getObject(tgtValue, body, err)
}

func postObject(tgtValue interface{}, endpoint string, params map[string]string, bodyType string, body []byte, auths ...Authentication) error {
	body, err := postBody(endpoint, params, bodyType, body, auths...)
	return _getObject(tgtValue, body, err)
}

func _getObject(tgtValue interface{}, body []byte, err error) error {
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, &tgtValue); err != nil {
		return err
	}

	//logg.Debug("getJSON: ", tgtValue, string(body))
	return nil
}

func getErrorMessage(resp *http.Response) ErrorMessage {
	var contentType = strings.SplitN(resp.Header.Get("Content-type"), ";", 2)
	var body []byte
	var err error
	var rc ErrorMessage

	if len(contentType) < 2 {
		logg.Fatal("missing/malformed response content type: ", resp.Header.Get("Content-type"))
	}

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		logg.Fatal("Failed to read response body: ", err)
	}

	switch contentType[0] {
	case "text/plain":
		rc.Code = resp.StatusCode
		rc.Status = "error"
		rc.Msg = string(body)
	case "application/json":
		if err = json.Unmarshal(body, &rc); err != nil {
			logg.Infof("response body: '%s'", string(body))
			logg.Fatalf("Failed to parse response message (response: %s): %s", resp.Status, err)
		}
	default:
		logg.Fatal("Unexpected error response format: ", resp.Header.Get("Content-type"))
	}

	return rc
}
