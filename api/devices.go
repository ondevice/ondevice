package api

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/ondevice/ondevice/logg"
)

// Device -- state info for a specific device
type Device struct {
	ID        string `json:"id"`
	CreatedAt int64  `json:"createdAt"`
	IP        string `json:"ip,omitempty"`
	State     string `json:"state"`
	StateTs   int64  `json:"stateTs,omitempty"`
	Name      string `json:"name,omitempty"`
	Version   string `json:"version,omitempty"`

	Active *bool                  `json:"active,omitempty"`
	Parent string                 `json:"parent,omitempty"`
	Props  map[string]interface{} `json:"props,omitempty"`
}

type deviceResponse struct {
	Devices []Device
}

type propertyListResponse struct {
	Props map[string]map[string]interface{}
}

// DeleteDevice -- remove a device (will fail if the device has recently been online)
func DeleteDevice(devID string, auths ...Authentication) error {
	var _, err = deleteBody("/device/"+devID, nil, "", nil, auths...)
	return err
}

// ListDevices -- list your devices and their state
func ListDevices(state string, props bool, auths ...Authentication) ([]Device, error) {
	var resp deviceResponse
	params := map[string]string{}

	if props {
		params["props"] = "true"
	}

	if err := getObject(&resp, "/devices", params, auths...); err != nil {
		return nil, err
	}

	// apply state filter
	rc := resp.Devices
	if state != "" {
		rc = []Device{}
		for _, dev := range resp.Devices {
			if dev.State == state {
				rc = append(rc, dev)
			}
		}
	}
	return rc, nil
}

// ListProperties -- Query device properties
func ListProperties(devID string, auths ...Authentication) (map[string]interface{}, error) {
	var rc propertyListResponse
	err := getObject(&rc, "/device/"+devID+"/props", nil, auths...)
	return _propertyList(rc, err)
}

// RemoveProperties -- remove one or more device properties
func RemoveProperties(devID string, props []string, auths ...Authentication) (map[string]interface{}, error) {
	var rc propertyListResponse
	values := url.Values{}

	if len(props) == 0 {
		return nil, fmt.Errorf("Can't delete empty list of properties")
	}

	for _, prop := range props {
		values.Add(prop, "")
	}

	obj := map[string][]string{"props": props}
	data, _ := json.Marshal(obj)

	err := deleteObject(&rc, "/device/"+devID+"/props", nil, "application/json", data, auths...)
	return _propertyList(rc, err)
}

// SetProperties -- Set device property values
func SetProperties(devID string, props map[string]string, auths ...Authentication) (map[string]interface{}, error) {
	var rc propertyListResponse
	values := url.Values{}

	if len(props) == 0 {
		return nil, fmt.Errorf("Can't set empty list of properties")
	}

	for k, v := range props {
		values.Add(k, v)
	}

	// TODO use the JSON request here
	err := postObject(&rc, "/device/"+devID+"/props", nil, "application/x-www-form-urlencoded", []byte(values.Encode()), auths...)
	return _propertyList(rc, err)
}

func _propertyList(data propertyListResponse, err error) (map[string]interface{}, error) {
	if err != nil {
		logg.Fatal("Couldn't get device properties: ", err)
	}

	// simply return the first element
	for _, prop := range data.Props {
		return prop, nil
	}

	return nil, fmt.Errorf("Got empty property response")
}
