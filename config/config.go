package config

import (
	"os"
	"os/user"
	"path/filepath"

	"path"

	"github.com/ondevice/ondevice/logg"
	"gopkg.in/ini.v1"
)

// if not nil, this will be used instead of `~/.config/ondevice` (mainly used for testing)
var _configPathOverride *string

var _fileOverrides = map[string]string{}

func setConfigPath(path string) {
	_configPathOverride = &path
}

// GetConfigPath -- Return the full path of a file in our config directory (usually ~/.config/ondevice/)
// Can be overridden using setConfigPath() (for testing only) or SetFilePath()
func GetConfigPath(filename string) string {
	if _, ok := _fileOverrides[filename]; ok {
		return _fileOverrides[filename]
	}

	// global config path override (used in unit tests)
	// TODO replace with single file overrides
	if _configPathOverride != nil {
		return path.Join(*_configPathOverride, filename)
	}

	var u, err = user.Current()
	if err != nil {
		logg.Fatal("Couldn't get current user", err)
	}

	return path.Join(u.HomeDir, ".config/ondevice", filename)
}

// GetVersion -- Returns the app version
func GetVersion() string {
	return "0.3.9"
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

// SetFilePath -- Override a config file's path (e.g. to satisfy standard OS paths)
func SetFilePath(filename string, path string) {
	_fileOverrides[filename] = path
}

// SetValue -- create/update a config value
func SetValue(section string, key string, value string) error {
	path := GetConfigPath("ondevice.conf")

	cfg, err := ini.InsensitiveLoad(path)
	if err != nil {
		return err
	}

	s, err := cfg.GetSection(section)
	if err != nil {
		return err
	}

	k, err := s.GetKey(key)
	if err != nil {
		return err
	}

	k.SetValue(value)

	tmpPath := filepath.Join(filepath.Dir(path), ".ondevice.conf.tmp")
	if err = cfg.SaveTo(tmpPath); err != nil {
		return err
	}
	if err = os.Rename(tmpPath, path); err != nil {
		return err
	}

	return nil
}
