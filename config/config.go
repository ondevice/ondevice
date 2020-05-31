package config

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"

	"path"

	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// if not nil, this will be used instead of `~/.config/ondevice` (mainly used for testing)
var _configPath string

var _fileOverrides = map[string]string{}

var version = "0.0.1-devel"

// GetConfigPath -- Return the full path of a file in our config directory (usually ~/.config/ondevice/)
// Can be overridden using setConfigPath() (for testing only) or SetFilePath()
func GetConfigPath(filename string) string {
	if _, ok := _fileOverrides[filename]; ok {
		return _fileOverrides[filename]
	}

	// global config path override (used in unit tests)
	// TODO replace with single file overrides
	if _configPath != "" {
		return path.Join(_configPath, filename)
	}

	var u, err = user.Current()
	var homeDir string

	if err == nil {
		homeDir = u.HomeDir
	} else {
		// This can happen when cross-compiling (it crept up on build-linux-armhf
		// even though https://github.com/golang/go/issues/14626 has been closed for quite some time now)
		//logrus.WithError(err).Warning("failed to get user.Current(), using $HOME instead")
		homeDir = os.Getenv("HOME")
		if homeDir == "" {
			logrus.Fatal("Couldn't get current user (and $HOME is empty): ", err)
		}
	}

	return path.Join(homeDir, ".config/ondevice", filename)
}

// GetInt -- Returns the specified integer config value (or defaultValue if not found or on error)
func GetInt(section string, key string, defaultValue int) int {
	var val, err = GetValue(section, key)
	if err != nil {
		return defaultValue
	}

	rc, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		logrus.WithError(err).Warningf("error parsing '%s.%s'", section, key)
		return defaultValue
	}

	return int(rc)
}

// GetVersion -- Returns the app version
func GetVersion() string {
	return version
}

// GetValue -- Get a configuration value (as string)
func GetValue(section string, key string) (string, error) {
	path := GetConfigPath("ondevice.conf")

	cfg, err := ini.InsensitiveLoad(path)
	if err != nil {
		return "", err
	}

	s, err := cfg.GetSection(section)
	if err != nil {
		return "", err
	}

	val, err := s.GetKey(key)
	if err != nil {
		return "", err
	}

	return val.String(), nil
}

// SetAuth -- Set client/user authentication details
func SetAuth(scope, user, auth string) error {
	if scope != "client" && scope != "device" {
		// panic instead of returning an error (since it pretty much has to be a programming error)
		logrus.Fatal("config.SetAuth(): scope needs to be one of 'device' and 'client': ", scope)
	}

	if err := SetValue(scope, "user", user); err != nil {
		return err
	}

	if err := SetValue(scope, "auth", auth); err != nil {
		return err
	}

	return nil
}

// SetFilePath -- Override a config file's path (e.g. to satisfy standard OS paths)
func SetFilePath(filename string, path string) {
	_fileOverrides[filename] = path
}

// SetValue -- create/update a config value
func SetValue(section string, key string, value string) error {
	path := GetConfigPath("ondevice.conf")

	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return err
	}

	cfg, err := ini.InsensitiveLoad(path)
	if os.IsNotExist(err) {
		logrus.Debug("creating new ondevice.conf")
		cfg = ini.Empty()
	} else if err != nil {
		return err
	}

	s := cfg.Section(section)
	k := s.Key(key)

	k.SetValue(value)

	// save to a temporary file and only replace the old file if successful
	// (to avoid corrupting the config file)
	tmpPath := filepath.Join(filepath.Dir(path), ".ondevice.conf.tmp")
	if err = cfg.SaveTo(tmpPath); err != nil {
		return err
	}
	if err = os.Chmod(tmpPath, 0600); err != nil {
		return err
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return err
	}

	return nil
}
