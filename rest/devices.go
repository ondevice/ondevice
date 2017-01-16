package rest

// Device -- state info for a specific device
type Device struct {
	ID      string                 `json:"id"`
	IP      string                 `json:"ip,omitempty"`
	State   string                 `json:"state"`
	StateTs int                    `json:"stateTs,omitempty"`
	Name    string                 `json:"name,omitempty"`
	Version string                 `json:"version,omitempty"`
	Props   map[string]interface{} `json:"props,omitempty"`
}

type deviceResponse struct {
	Devices []Device
}

// ListDevices -- list your devices and their state
func ListDevices(props bool, auths ...Authentication) ([]Device, error) {
	var resp deviceResponse
	params := map[string]string{}

	if props {
		params["props"] = "true"
	}

	if err := getObject(&resp, "/devices", params, auths...); err != nil {
		return nil, err
	}

	return resp.Devices, nil
}
