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
func (c Config) GetInt(key configKey) int {
	if s := c.cfg.Section(key.section); s != nil {
		if k := s.Key(key.key); k != nil {
			if rc, err := k.Int(); err == nil {
				return rc
			}
		}
	}

	var rc, err = strconv.ParseInt(key.defaultValue, 10, 32)
	if err != nil {
		logrus.WithError(err).Fatalf("failed to parseInt for config key '%v': '%s'", key, key.defaultValue)
		return 0
	}

	return int(rc)
}

// GetString -- Get a configuration value (as string)
func (c Config) GetString(section string, key string) (string, error) {
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

// HasKey -- returns true iff the given value has been defined (defaults don't count)
func (c Config) HasKey(section string, key string) bool {
	if s := c.cfg.Section(section); s != nil {
		if k := s.Key(key); k != nil {
			return true
		}
	}

	return false
}

// IsChanged -- returns true once SetValue() has been called
func (c Config) IsChanged() bool { return c.changed }

// SetValue -- create/update a config value - don't forget to call Write() afterwards
func (c Config) SetValue(section string, key string, value string) error {
	var s *ini.Section
	var err error

	c.changed = true

	if s = c.cfg.Section(section); s == nil {
		if s, err = c.cfg.NewSection(section); err != nil {
			logrus.WithError(err).Errorf("failed to create config section: '%s'", section)
			return err
		}
	}

	var k *ini.Key
	if k = s.Key(key); k == nil {
		if k, err = s.NewKey(key, value); err != nil {
			logrus.WithError(err).Errorf("failed to create new config value '%s.%s'='%s'", section, key, value)
			return err
		}
		return nil
	}

	k.SetValue(value)
	return nil
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

// GetFilePath -- returns the path of a file, relative to ondevice.conf
func (c Config) GetFilePath(filename string) string {
	var dir = filepath.Dir(c.path)
	return filepath.Join(dir, filename)
}

// GetVersion -- Returns the app version
func GetVersion() string {
	return version
}

// LoadAuth -- fetches information stored in auth.json
//
// uses [path].auth_json as reference
func (c Config) LoadAuth() internal.AuthJSON {
	var path = c.GetFilePath("auth.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		c.migrateAuth()
	}

	return internal.ReadAuth(path)
}

// SetAuth -- Set client/user authentication details
func SetAuth(scope, user, auth string) error {
	if scope != "client" && scope != "device" {
		// panic instead of returning an error (since it pretty much has to be a programming error)
		logrus.Fatal("config.SetAuth(): scope needs to be one of 'device' and 'client': ", scope)
	}

	var cfg, err = Load()
	if err != nil {
		logrus.WithError(err).Error("failed to read ondevice.conf")
		return err
	}
	if err := cfg.SetValue(scope, "user", user); err != nil {
		return err
	}

	if err := cfg.SetValue(scope, "auth", auth); err != nil {
		return err
	}

	return cfg.Write()
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
