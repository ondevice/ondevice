package internal

import (
	"fmt"
	"strings"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// DeviceListCompletion -- configurable device list shell completion
type DeviceListCompletion struct {
	// DontIgnoreUser -- unless this is set, we'll ignore everything including the first '@' (i.e. the user part in 'user@devId')
	DontIgnoreUser bool
}

// Run -- proviceds shell completion for devIDs
//
func (c DeviceListCompletion) Run(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// TODO cache the results (e.g. in /tmp)
	// TODO we're using the term 'user' for both of "sshUser@ondeviceUser.devId" here...

	// we'll compile a list of unqualified and qualified devIDs for the main user as well as the list of other users (ending with a dot)
	// e.g.: 'a me.a you.'
	// if toComplete contains a dot, only the user in question will be queried
	var matchingDevices []string
	var rc = cobra.ShellCompDirectiveNoFileComp
	var prefix string

	prefix, toComplete = c.stripUserAt(toComplete)
	var dotPos = strings.Index(toComplete, ".")

	if dotPos < 1 { // no dot in the hostname part
		var auth api.Authentication
		var err error
		if auth, err = api.GetClientAuth(); err != nil {
			logrus.Fatal("missing client auth, have you run 'ondevice login'?")
			return nil, cobra.ShellCompDirectiveError
		}

		var allDevices []api.Device
		if allDevices, err = api.ListDevices("", false, auth); err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		// unqualified devIDs
		for _, dev := range allDevices {
			if strings.HasPrefix(dev.UnqualifiedID(), toComplete) {
				matchingDevices = append(matchingDevices, prefix+dev.UnqualifiedID())
			}
		}

		// qualified devIDs
		for _, dev := range allDevices {
			if strings.HasPrefix(dev.ID, toComplete) {
				matchingDevices = append(matchingDevices, prefix+dev.ID)
			}
		}

		for _, clientUser := range config.ListAuthenticatedUsers() {
			if strings.HasPrefix(clientUser, toComplete) {
				matchingDevices = append(matchingDevices, fmt.Sprintf("%s%s.", prefix, clientUser))
			}
		}
	} else { // user.devId
		var username = toComplete[:dotPos]
		var auth api.Authentication
		var err error
		if auth, err = api.GetClientAuthForUser(username); err != nil {
			logrus.Fatalf("missing client auth for user '%s'", username)
			return nil, cobra.ShellCompDirectiveError
		}

		var allDevices []api.Device
		if allDevices, err = api.ListDevices("", false, auth); err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		// qualified devIDs
		for _, dev := range allDevices {
			if strings.HasPrefix(dev.ID, toComplete) {
				matchingDevices = append(matchingDevices, prefix+dev.ID)
			}
		}
	}

	return matchingDevices, rc
}

// stripUserAt -- if DontIgnoreUser is false and userAtHost contains an '@', return the part after the at
func (c DeviceListCompletion) stripUserAt(userAtHost string) (userAt string, hostname string) {
	if c.DontIgnoreUser {
		return "", userAtHost
	}

	var atIndex = strings.Index(userAtHost, "@")
	if atIndex >= 0 {
		return userAtHost[:atIndex+1], userAtHost[atIndex+1:]
	}
	return "", userAtHost
}
