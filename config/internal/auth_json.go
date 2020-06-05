package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// ErrNoClientAuth -- returned if we lack client credentials
var ErrNoClientAuth = errors.New("missing client auth, try 'ondevice login'")

// ErrNoDeviceAuth -- returned if we lack client credentials
var ErrNoDeviceAuth = errors.New("missing device auth, try 'ondevice login'")

// AuthJSON -- marshals/unmarshals the contents of the auth.json file
type AuthJSON struct {
	Client       AuthEntry
	Device       AuthEntry
	ExtraClients []AuthEntry `json:",omitempty"`

	path string
	err  error

	isChanged bool
}

// GetClientAuth -- returns client credentials
func (j AuthJSON) GetClientAuth() (Auth, error) {
	if j.err != nil {
		return nil, j.err
	}
	if j.Client.UserField == "" || j.Client.KeyField == "" {
		return nil, ErrNoClientAuth
	}
	return j.Client, nil
}

// GetClientAuthForDevice -- Returns credentials for the given devID
//
// with unqualified devIDs, this will do the same as GetClientAuth().
// But if the devID has a user prefix (and we have extra credentials for that user), it'll return those instead
func (j AuthJSON) GetClientAuthForDevice(devID string) (Auth, error) {
	if strings.Contains(devID, ".") {
		parts := strings.SplitN(devID, ".", 2)
		if auth, err := j.GetClientAuthForUser(parts[0]); err == nil {
			return auth, nil
		}
	}

	return j.GetClientAuth()
}

// GetClientAuthForUser -- get the authentication credentials for a specific client user
func (j AuthJSON) GetClientAuthForUser(username string) (Auth, error) {
	var auth, err = j.GetClientAuth()
	if err == nil && strings.ToLower(auth.User()) == strings.ToLower(username) {
		return auth, nil
	}

	for _, auth = range j.ExtraClients {
		if strings.ToLower(auth.User()) == strings.ToLower(username) {
			return auth, nil
		}
	}

	return nil, fmt.Errorf("no client credentials found for user '%s'", username)
}

// GetDeviceAuth -- Get the device authentication
func (j AuthJSON) GetDeviceAuth() (Auth, error) {
	if j.err != nil {
		return nil, j.err
	}
	if j.Device.UserField == "" || j.Device.KeyField == "" {
		return nil, ErrNoDeviceAuth
	}
	return j.Device, nil
}

// GetDeviceKey -- returns the unique deviceKey that identifies this device (or "")
func (j AuthJSON) GetDeviceKey() string {
	// won't check j.err here
	return j.Client.DeviceKey
}

func (j AuthJSON) IsChanged() bool { return j.isChanged }

// ListClientUsers -- returns the names of users we have client auth for
func (j AuthJSON) ListClientUsers() []string {
	// TODO this is messy but will do for now - we'll improve this once we have a separate auth file

	var rc []string
	var uniqueUsers = make(map[string]bool)

	if mainUser := j.Client.UserField; mainUser != "" {
		rc = append(rc, mainUser)
		uniqueUsers[strings.ToLower(mainUser)] = true
	}

	for _, auth := range j.ExtraClients {
		var lowerName = strings.ToLower(auth.UserField)
		if !uniqueUsers[lowerName] {
			rc = append(rc, auth.UserField)
			uniqueUsers[lowerName] = true
		}
	}

	return rc
}

// SetClientAuth -- updates the client credentials (don't forget to call .Write())
func (j AuthJSON) SetClientAuth(user string, key string) {
	j.Client.UserField = user
	j.Client.KeyField = key
	j.isChanged = true
}

// SetDeviceAuth -- updates the device credentials (don't forget to call .Write())
func (j AuthJSON) SetDeviceAuth(user string, key string) {
	j.Device.UserField = user
	j.Device.KeyField = key
	j.isChanged = true
}

// Write -- atomically update auth.json
func (j AuthJSON) Write() error {
	var data, err = json.Marshal(j)
	if err != nil {
		logrus.WithError(err).Error("failed to marshal auth.json data")
		return err
	}

	return WriteFile(data, j.path, 0o600)
}

// ReadAuth -- Read auth.json from the given file path
//
// On errors, this will silently return empty credential information - but calls to AuthJSON methods will fail returning the error we suppressed
func ReadAuth(path string) AuthJSON {
	var file *os.File
	var err error
	var rc = AuthJSON{
		path: path,
	}

	if file, err = os.Open(path); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to open auth.json")
		rc.err = err
		return rc
	}

	var stat os.FileInfo
	if stat, err = file.Stat(); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to check auth.json file mode")
		rc.err = err
		return rc
	}

	if stat.Mode() != 0o600 {
		// Note: this may fail on certain platforms - TODO add special cases as they happen
		logrus.WithField("path", path).Fatalf("auth.json isn't supposed to be accessible by other users (mode: %o)", stat.Mode())
	}

	var data []byte
	if data, err = ioutil.ReadAll(file); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to read auth.json")
		rc.err = err
		return rc
	}

	if err = json.Unmarshal(data, &rc); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to parse auth.json")
		rc.err = err
		return rc
	}

	// legacy overrides (probably will be deleted in v1.0)
	if user, key := os.Getenv("ONDEVICE_USER"), os.Getenv("ONDEVICE_AUTH"); user != "" || key != "" {
		logrus.Warning("ONDEVICE_USER/ONDEVICE_AUTH overrides present. these will become obsolete soon")
		rc.Client = AuthEntry{
			UserField: os.Getenv("ONDEVICE_USER"),
			KeyField:  os.Getenv("ONDEVICE_AUTH"),
		}
		rc.Device = AuthEntry{
			UserField: os.Getenv("ONDEVICE_USER"),
			KeyField:  os.Getenv("ONDEVICE_AUTH"),
			DeviceKey: rc.Device.DeviceKey,
		}
		rc.err = nil
	}

	return rc
}
