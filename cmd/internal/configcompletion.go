package internal

import (
	"strings"

	"github.com/ondevice/ondevice/config"
	"github.com/spf13/cobra"
)

// ConfigCompletion -- configurable config key shell completion
type ConfigCompletion struct {
	// WithReadOnly -- if set, also include read-only keys in the results
	WithReadOnly bool
	Suffix       string
}

// Run -- proviceds shell completion for devIDs
//
func (c ConfigCompletion) Run(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO cache the results (e.g. in /tmp)
	// TODO we're using the term 'user' for both of "sshUser@ondeviceUser.devId" here...

	// we'll compile a list of unqualified and qualified devIDs for the main user as well as the list of other users (ending with a dot)
	// e.g.: 'a me.a you.'
	// if toComplete contains a dot, only the user in question will be queried
	var matchingKeys []string
	var rc = cobra.ShellCompDirectiveNoFileComp

	for k := range config.AllKeys(c.WithReadOnly) {
		if strings.HasPrefix(k, toComplete) {
			matchingKeys = append(matchingKeys, k+c.Suffix)
		}
	}

	return matchingKeys, rc
}
