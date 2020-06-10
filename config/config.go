package config

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/mitchellh/go-homedir"
	"github.com/ondevice/ondevice/config/internal"
	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// if not empty, this will be used instead of `~/.config/ondevice/ondevice.conf`
var _configPath string

var version = "0.0.1-devel"

// Config -- config file's contents, acquired using config.Read()
type Config struct {
	cfg  *ini.File
	path string

	changed bool // set by SetValue()
}

// Load -- fetches the contents of ondevice.conf
func Load() (Config, error) {
	var rc Config
	var err error

	// fetch path
	if _configPath != "" {
		rc.path = _configPath
	} else {
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
				logrus.WithError(err).Fatal("couldn't get current user (and $HOME is empty)")
			}
		}

		rc.path = path.Join(homeDir, ".config/ondevice/ondevice.conf")
	}

	// read file
	if rc.cfg, err = ini.InsensitiveLoad(rc.path); err != nil {
		if !os.IsNotExist(err) {
			logrus.WithError(err).Error("config.Read(): file not found: ondevice.conf")
		}
		rc.cfg = ini.Empty()
		return rc, err
	}
	return rc, nil
}

// MustLoad -- calls Load() and fatals on error (except on os.IsNotExist - in which case a zero Config struct will be returned)
func MustLoad() Config {
	var rc, err = Load()
	if os.IsNotExist(err) {
		return rc
	} else if err != nil {
		logrus.WithError(err).Fatal("failed to load ondevice.conf")
		panic("should have fataled by now")
	}
	return rc
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
func (c Config) GetInt(key Key) int {
	var strVal = c.GetString(key)
	var rc, err = strconv.ParseInt(strVal, 0, 32)
	if err == nil {
		return int(rc)
	}

	// -> err != nil - assume it's the user's value
	//  if empty, simply return the default value
	//  if not, log the error and also return the default value
	if strVal != "" {
		logrus.WithError(err).Errorf("failed to parse value for '%v' (expected integer): '%s'", key, strVal)
	}
	rc, err = strconv.ParseInt(key.defaultValue, 10, 32)
	if err != nil {
		// fail hard because the default value is not an int (i.e. there's a coding issue)
		logrus.WithError(err).Fatalf("expected integer default value for config key '%v', not '%s'", key, key.defaultValue)
		return 0
	}

	return int(rc)
}

// GetString -- Fetch a configuration value (will return key.defaultValue if not defined)
func (c Config) GetString(key Key) string {
	if s, err := c.cfg.GetSection(key.section); err == nil {
		if k, err := s.GetKey(key.key); err == nil {
			return k.String()
		}
	}
	return key.defaultValue
}

// HasKey -- returns true iff the given value has been defined (defaults don't count)
func (c Config) HasKey(section string, key string) bool {
	if s, err := c.cfg.GetSection(section); err == nil {
		if _, err := s.GetKey(key); err == nil {
			return true
		}
	}

	return false
}

// IsChanged -- returns true once SetValue() has been called
func (c Config) IsChanged() bool { return c.changed }

// SetNX -- set a config value, but only if it hasn't been defined before
func (c *Config) SetNX(key Key, value string) error {
	if c.HasKey(key.section, key.key) {
		logrus.Infof("not setting %q=%q, already set", key, value)
		return nil
	}
	return c.SetValue(key, value)
}

// SetValue -- create/update a config value - don't forget to call Write() afterwards
func (c *Config) SetValue(key Key, value string) error {
	var s *ini.Section
	var err error

	c.changed = true

	if s = c.cfg.Section(key.section); s == nil {
		if s, err = c.cfg.NewSection(key.section); err != nil {
			logrus.WithError(err).Errorf("failed to create config section: '%s'", key.section)
			return err
		}
	}

	var k *ini.Key
	if k = s.Key(key.key); k == nil {
		if k, err = s.NewKey(key.key, value); err != nil {
			logrus.WithError(err).Errorf("failed to create new config value '%v'='%s'", key, value)
			return err
		}
		return nil
	}

	if err = key.Validate(value); err != nil {
		logrus.WithError(err).Errorf("failed to set '%v' to '%s': validation failed", key, value)
		return err
	}

	k.SetValue(value)
	return nil
}

// Unset -- reverts the given config key to its default value
//
// (if the key/section in question doesn't exist nothing happens)
func (c *Config) Unset(key Key) {
	if s, err := c.cfg.GetSection(key.section); err == nil {
		s.DeleteKey(key.key)
	}
}

// Write -- writes ondevice.conf (using writeFile() for safe replacement)
//
// Make sure to only write freshly read Config files. otherwise you'll dramatically
// increase the likelihood of creating race conditions
func (c Config) Write() error {
	var buff bytes.Buffer
	if _, err := c.cfg.WriteTo(&buff); err != nil {
		logrus.WithError(err).Error("failed to write ondevice.conf")
		return err
	}
	return internal.WriteFile(buff.Bytes(), c.path, 0o644)
}

// GetFilePath -- returns the path of the file in question, relative to ondevice.conf
func (c Config) GetFilePath(key Key) string {
	// unlike python's path.join() go's version will always concat the two!?!
	var rc = c.GetString(key)
	if !filepath.IsAbs(rc) {
		var dir = filepath.Dir(c.path)
		rc = filepath.Join(dir, c.GetString(key))
	}
	return rc
}

// GetVersion -- Returns the app version
func GetVersion() string {
	return version
}

// LoadAuth -- fetches information stored in auth.json
//
// uses [path].auth_json as reference
func (c Config) LoadAuth() AuthConfig {
	var path = c.GetFilePath(PathAuthJSON)
	var rc = internal.LoadAuth(path)
	if rc.Error() != nil && os.IsNotExist(rc.Error()) {
		var err error
		if rc, err = c.migrateAuth(path); err != nil {
			logrus.WithError(err).Fatal("failed to migrate credentials to auth.json")
		}
	}

	return &rc
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

	// TODO we're always reading the configuration here -> think about caching this
	//   Note: for ondevice daemon (and other long-running commands) the config should be re-read periodically
	var cfg Config
	if cfg, err = Load(); err != nil {
		logrus.WithError(err).Error("failed to read ondevice.conf")
		return
	}

	// set a default timeout of 30sec for REST API calls (will be reset in long-running commands)
	// TODO use a builder pattern to be able to specify this on a per-request basis
	// Note: doesn't affect websocket connections
	var timeout = time.Duration(cfg.GetInt(KeyClientTimeout))
	http.DefaultClient.Timeout = timeout * time.Second
}
