package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/ondevice/ondevice/config/internal"
	"github.com/sirupsen/logrus"
)

// migrateAuth -- move stuff from ondevice.conf to auth.json
func (c Config) migrateAuth() error {
	if !c.HasKey("device", "auth") && !c.HasKey("client", "auth") {
		return nil // doesn't have auth
	}
	logrus.Info("migrating auth info to auth.json")
	// TODO remove old auth once we're confident things work as expected

	// target configurations
	var auth internal.AuthJSON
	var err error

	// chain getValue calls until the first error happened
	var getValue = func(olderr error, section, key string) (string, error) {
		if olderr != nil {
			return "", olderr
		}
		return c.GetStringOld(section, key)
	}

	var getOrDefault = func(section, key, defaultValue string) string {
		if s := c.cfg.Section(section); s != nil {
			if k := s.Key(key); k != nil {
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

	auth.Client.UserField, err = getValue(err, "client", "user")
	auth.Client.KeyField, err = getValue(err, "client", "auth")
	if err != nil {
		logrus.WithError(err).Fatal("failed to convert client config")
		return err
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
