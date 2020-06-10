package control

import (
	"encoding/json"
	"errors"
	"net/http"
)

// response -- wraps http.Response
type response struct {
	resp *http.Response
	err  error
}

// Close -- cleans up response resources
func (r response) Close() error {
	if r.resp != nil && r.resp.Body != nil {
		return r.resp.Body.Close()
	}
	return r.err
}

// Error -- returns any errors happening when running the request
func (r response) Error() error {
	return r.err
}

// ParseJSON -- parses the response as JSON and closes the response body
func (r response) ReadJSON(tgt interface{}) error {
	if r.resp == nil {
		return errors.New("request failed, can't unmarshal response")
	}
	defer r.resp.Body.Close()

	var decoder = json.NewDecoder(r.resp.Body)
	return decoder.Decode(&tgt)
}
