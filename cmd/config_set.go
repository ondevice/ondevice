package cmd

import (
	"os"
	"strings"

	"github.com/ondevice/ondevice/cmd/internal"
	"github.com/ondevice/ondevice/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	var c configSetCmd
	c.Command = cobra.Command{
		Use:   "set <key>=<value>...",
		Short: "set one or more configuration values",
		Long: `ondevice config set updates one or more configuration values.
	If you specify the same key more than once, the last one in the list wins`,
		Example: `  $ ondevice config set ssh.path=/usr/local/bin/ssh rsync.path=echo
	  $ ondevice config set client.timeout=5`,
		Run:               c.Run,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: internal.ConfigCompletion{WithReadOnly: false, Suffix: "="}.Run,
	}

	configCmd.AddCommand(&c.Command)
	c.Flags().BoolVar(&c.noOverwriteFlag, "no-overwrite", false, "if set, won't overwrite an already existing value (i.e. only set if not yet defined')")
	c.Flags().BoolVar(&c.dryRunFlag, "dry-run", false, "if set, only does validation but won't update the config file")
}

type configSetCmd struct {
	cobra.Command

	noOverwriteFlag bool
	dryRunFlag      bool
}

func (c *configSetCmd) Run(cmd *cobra.Command, args []string) {
	var rc = 0

	var cfg = config.MustLoad()
	for _, keyValue := range args {
		var parts = strings.SplitN(keyValue, "=", 2)
		if len(parts) != 2 {
			logrus.Fatalf("malformed argument, expected key=value pairs: '%s'", keyValue)
		}
		var key = config.FindKey(parts[0])
		var newValue = parts[1]

		if key == nil {
			logrus.Fatalf("config key not found: '%s'", keyValue)
			rc = 1
			continue
		}

		if c.noOverwriteFlag {
			if err := cfg.SetNX(*key, newValue); err != nil {
				rc = 1
			}
		} else {
			if err := cfg.SetValue(*key, newValue); err != nil {
				rc = 1
			}
		}
	}

	if rc != 0 {
		os.Exit(rc)
	}
	if !cfg.IsChanged() {
		logrus.Info("ondevice set: nothing changed")
		return
	}

	if !c.dryRunFlag {
		if err := cfg.Write(); err != nil {
			logrus.WithError(err).Error("failed to write config")
		}
	}
}
