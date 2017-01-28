package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

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
		logg.Fatal("Failed request", err)
	}
	req.Header.Add("Authorization", auth.GetAuthHeader())
	req.Header.Add("User-agent", fmt.Sprintf("ondevice v%s", config.GetVersion()))

	if body != nil {
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
		// TODO make the content type configurable
		req.Header.Add("Content-type", bodyType)
	}

	client := http.Client{}
	resp, err := client.Do(req)

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

	if resp.StatusCode == 401 {
		return nil, fmt.Errorf("Authentication failed!")
	} else if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP error while querying key info: %d %s", resp.StatusCode, resp.Status)
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
