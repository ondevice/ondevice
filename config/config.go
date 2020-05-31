package config

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// if not empty, this will be used instead of `~/.config/ondevice/ondevice.conf`
var _configPath string

var version = "0.0.1-devel"

// Config -- config file's contents, acquired using config.Read()
type Config struct {
	cfg *ini.File
}

// Read -- fetches the contents of ondevice.conf
func Read() (Config, error) {
	var rc Config
	var err error
	path := GetConfigPath("ondevice.conf")

	if rc.cfg, err = ini.InsensitiveLoad(path); err != nil {
		return rc, err
	}
	return rc, nil
}

// AllValues -- returns a flattened key/value dictionary for all values in ondevice.conf
func (c Config) AllValues() map[string]string {
	var rc = make(map[string]string)

	for _, s := range c.cfg.Sections() {
		for _, k := range s.Keys() {
			var key = fmt.Sprintf("%s.%s", s.Name(), k.Name())
			var value = k.String()
			rc[key] = value
		}
	}

	return rc
}

// GetInt -- Returns the specified integer config value (or defaultValue if not found or on error)
func (c Config) GetInt(section string, key string, defaultValue int) int {
	var s = c.cfg.Section(section)
	if s != nil {
		return defaultValue // missing section
	}

	var k = s.Key(key)
	if k != nil {
		return defaultValue // missing key
	}

	var rc, err = k.Int()
	if err != nil {
		logrus.WithError(err).Errorf("expected integer value for config key '%s.%s'", section, key)
		return defaultValue
	}

	return rc
}

// GetString -- Get a configuration value (as string)
func (c Config) GetString(section string, key string) (string, error) {
	if c.cfg == nil {
		return "", errors.New("trying to read from nil config")
	}
	s, err := c.cfg.GetSection(section)
	if err != nil {
		return "", err
	}

	val, err := s.GetKey(key)
	if err != nil {
		return "", err
	}

	return val.String(), nil
}

// GetConfigPath -- Return the full path of a file in our config directory (usually ~/.config/ondevice/)
// Can be overridden using setConfigPath() (for testing only) or SetFilePath()
func GetConfigPath(filename string) string {
	// global config path override (used in unit tests)
	// TODO replace with single file overrides
	if _configPath != "" {
		return path.Join(filepath.Dir(_configPath), filename)
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

// GetVersion -- Returns the app version
func GetVersion() string {
	return version
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

// Init -- sets up configuration, called by cobra.OnInitialize()
func Init(cfgFile string) {
	if cfgFile != "" {
		// Use config file from the flag.
		_configPath = cfgFile
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// TODO maybe use another path on windows (and other )
		_configPath = filepath.Join(home, ".config/ondevice/ondevice.conf")
	}

	// create parent directory
	var err error
	if err = os.MkdirAll(filepath.Dir(_configPath), 0o755); err != nil {
		logrus.WithError(err).Fatalf("failed to create config directory: '%s'", filepath.Dir(_configPath))
	}

	var cfg Config
	if cfg, err = Read(); err != nil {
		logrus.WithError(err).Error("failed to read ondevice.conf")
		return
	}

	// set a default timeout of 30sec for REST API calls (will be reset in long-running commands)
	// TODO use a builder pattern to be able to specify this on a per-request basis
	// Note: doesn't affect websocket connections
	var timeout = time.Duration(cfg.GetInt("client", "timeout", 30))
	http.DefaultClient.Timeout = timeout * time.Second
}
