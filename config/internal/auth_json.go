package internal

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/sirupsen/logrus"
)

// AuthJSON -- marshals/unmarshals the contents of the auth.json file
type AuthJSON struct {
	Client       AuthEntry
	Device       AuthEntry
	ExtraClients []AuthEntry `json:",omitempty"`
}

// ReadAuth -- Read auth.json from the given file path
//
// for now, this will migrate the old-style ondevice.conf auth info into the new auth.json
func ReadAuth(path string) (AuthJSON, error) {
	var file *os.File
	var err error
	var rc AuthJSON

	if file, err = os.Open(path); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to open auth.json")
		return rc, err
	}

	var stat os.FileInfo
	if stat, err = file.Stat(); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to check auth.json file mode")
		return rc, err
	}

	if stat.Mode() != 0o600 {
		// Note: this may fail on certain platforms - TODO add special cases as they happen
		logrus.WithField("path", path).Fatalf("auth.json isn't supposed to be accessible by other users (mode: %o)", stat.Mode())
	}

	var data []byte
	if data, err = ioutil.ReadAll(file); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to read auth.json")
		return rc, err
	}

	if err = json.Unmarshal(data, &rc); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to parse auth.json")
		return rc, err
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
	}

	return rc, nil
}
