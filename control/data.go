package control

// DeviceState -- A Device's state
type DeviceState struct {
	Version string            `json:"version"`
	Client  map[string]string `json:"client,omitempty"`
	Device  map[string]string `json:"device,omitempty"`
}
