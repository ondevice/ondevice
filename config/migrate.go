package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/ondevice/ondevice/config/internal"
	"github.com/sirupsen/logrus"
)

// migrateAuth -- move stuff from ondevice.conf to auth.json
//
// (up until v0.6.1 credentials were stored in ondevice.conf)
func (c Config) migrateAuth(path string) (internal.AuthJSON, error) {
	var rc internal.AuthJSON
	var err error

	var overrideUser = os.Getenv("ONDEVICE_USER")
	var overrideAuth = os.Getenv("ONDEVICE_AUTH")

	if !c.hasKey("device", "auth") && !c.hasKey("client", "auth") && overrideAuth == "" {
		return rc, nil // no credentials found, return empty AuthJSON instance
	}

	logrus.Info("migrating auth info to auth.json")
	// TODO remove old auth once we're confident things work as expected

	var getOrDefault = func(section, key, defaultValue string) string {
		if s, err := c.cfg.GetSection(section); err == nil {
			if k, err := s.GetKey(key); err == nil {
				return k.String()
			}
		}
		return defaultValue
	}

	// device auth
	rc.Device = internal.AuthEntry{
		UserField: getOrDefault("device", "user", ""),
		KeyField:  getOrDefault("device", "auth", ""),
		DeviceKey: getOrDefault("device", "key", ""),
	}

	// client section
	var extraClients []internal.AuthEntry
	for key, val := range c.AllValues() {
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
	rc.ExtraClients = extraClients

	rc.Client = internal.AuthEntry{
		UserField: getOrDefault("client", "user", ""),
		KeyField:  getOrDefault("client", "auth", ""),
	}

	if overrideUser != "" || overrideAuth != "" {
		// apply overrides
		if rc.Device.UserField != "" || rc.Device.KeyField != "" {
			// TODO think about how to behave here
			logrus.Fatal("migrateAuth(): got both device.user/device.auth and ONDEVICE_USER/ONDEVICE_AUTH")
		}

		// We'll only apply the overrides to the device credentials
		rc.Device.UserField = overrideUser
		rc.Device.KeyField = overrideAuth
	}

	var data []byte
	if data, err = json.Marshal(rc); err != nil {
		logrus.WithError(err).Fatal("failed to marshal auth.json")
		return rc.WithError(err), err
	}

	// TODO remove old auth from ondevice.conf

	// TODO use auth.Write() - that only works if auth.path is set correctly though
	if err = internal.WriteFile(data, path, 0o600); err != nil {
		logrus.WithError(err).Fatal("migrateConfig(): failed to write auth.json")
		return rc.WithError(err), err
	}

	return rc, nil
}
