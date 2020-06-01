package config

import (
	"encoding/json"
	"strings"

	"github.com/ondevice/ondevice/config/internal"
	"github.com/sirupsen/logrus"
)

// migrateAuth -- move stuff from ondevice.conf to auth.json
func migrateAuth() error {
	var cfg, err = Read()
	if err != nil {
		return err
	}

	// target configurations
	var auth internal.AuthJSON

	// chain getValue calls until the first error happened
	var getValue = func(olderr error, section, key string) (string, error) {
		if olderr != nil {
			return "", olderr
		}
		return cfg.GetString(section, key)
	}

	// device auth
	auth.Device.UserField, err = getValue(err, "device", "user")
	auth.Device.KeyField, err = getValue(err, "device", "auth")
	auth.Device.DeviceKey, err = getValue(err, "device", "key")
	if err != nil {
		logrus.WithError(err).Fatal("failed to convert device config")
		return err
	}

	// client section
	var extraClients []internal.AuthEntry
	for key, val := range cfg.AllValues() {
		if strings.HasPrefix(strings.ToLower(key), "client.auth_") {
			var username = key[len("client.auth_"):]

			if len(username) == 0 || len(val) == 0 {
				logrus.Fatalf("found empty auth data for key '%s'", key)
			}
			extraClients = append(extraClients, internal.AuthEntry{
				UserField: username,
				KeyField:  val,
			})
		}
	}
	auth.ExtraClients = extraClients

	auth.Client.UserField, err = getValue(err, "client", "user")
	auth.Client.KeyField, err = getValue(err, "client", "auth")
	if err != nil {
		logrus.WithError(err).Fatal("failed to convert client config")
		return err
	}

	// TODO apply overrides
	var data []byte
	if data, err = json.Marshal(auth); err != nil {
		logrus.WithError(err).Fatal("failed to marshal auth.json")
	}

	// TODO remove old auth from ondevice.conf

	// TODO use auth.Write()
	if err = internal.WriteFile(data, GetConfigPath("auth.json"), 0o600); err != nil {
		logrus.WithError(err).Fatal("migrateConfig(): failed to write auth.json")
		return err
	}

	return nil
}
