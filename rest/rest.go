package rest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Get -- issue a GET request to the API server
func get(endpoint string, params map[string]string, auths ...Authentication) (*http.Response, error) {
	var auth *Authentication

	if auths == nil {
		a, _ := CreateClientAuth()
		auth = &a
	} else {
		auth = &auths[0]
	}

	url := auth.getURL(endpoint)
	log.Print("GET request to URL ", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("Failed request", err)
	}
	req.Header.Add("Authorization", auth.getAuthHeader())

	client := http.Client{}
	resp, err := client.Do(req)

	return resp, err
}

func getBody(endpoint string, params map[string]string, auths ...Authentication) ([]byte, error) {
	resp, err := get(endpoint, params, auths...)
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

func getObject(tgtValue interface{}, endpoint string, params map[string]string, auths ...Authentication) error {
	body, err := getBody(endpoint, params, auths...)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, &tgtValue); err != nil {
		return err
	}

	//log.Println("getJSON: ", tgtValue, string(body))
	return nil
}
