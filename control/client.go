package control

// GetState -- Query the device state over the unix socket
func GetState() (DeviceState, error) {
	rc := DeviceState{}
	err := request{endpoint: "/state"}.Get().ReadJSON(&rc)
	return rc, err
}
