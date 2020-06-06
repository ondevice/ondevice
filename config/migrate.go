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
func (c Config) migrateAuth() error {
	if !c.HasKey("device", "auth") && !c.HasKey("client", "auth") {
		return nil // doesn't have auth
	}
	logrus.Info("migrating auth info to auth.json")
	// TODO remove old auth once we're confident things work as expected

	// target configurations
	var auth internal.AuthJSON
	var err error

	var getOrDefault = func(section, key, defaultValue string) string {
		if s, err := c.cfg.GetSection(section); err == nil {
			if k, err := s.GetKey(key); err == nil {
				return k.String()
			}
		}
		return defaultValue
	}

	// device auth
	auth.Device = internal.AuthEntry{
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
	auth.ExtraClients = extraClients

	auth.Client = internal.AuthEntry{
		UserField: getOrDefault("client", "user", ""),
		KeyField:  getOrDefault("client", "auth", ""),
	}

	if user, pw := os.Getenv("ONDEVICE_USER"), os.Getenv("ONDEVICE_AUTH"); user != "" || pw != "" {
		// apply overrides
		if auth.Device.UserField != "" || auth.Device.KeyField != "" {
			// TODO think about how to behave here
			logrus.Fatal("migrateAuth(): got both device.user/device.auth and ONDEVICE_USER/ONDEVICE_AUTH")
		}
		auth.Device.UserField = user
		auth.Device.KeyField = pw

		// TODO should we also apply them to .Client?
	}

	var data []byte
	if data, err = json.Marshal(auth); err != nil {
		logrus.WithError(err).Fatal("failed to marshal auth.json")
	}

	// TODO remove old auth from ondevice.conf

	// TODO use auth.Write()
	if err = internal.WriteFile(data, c.GetFilePath("auth.json"), 0o600); err != nil {
		logrus.WithError(err).Fatal("migrateConfig(): failed to write auth.json")
		return err
	}

	return nil
}
