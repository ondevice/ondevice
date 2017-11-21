package command

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

var scpFlags = sshParseFlags("12346BCpqrvc:F:i:l:o:P:S:")

func scpRun(args []string) int {
	scpPath := "/usr/bin/scp"

	args, opts := sshParseArgs(scpFlags, args)

	// parse all the args as possible remote files [[user@]devId]:/path/to/file
	for i := 0; i < len(args); i++ {
		var arg = args[i]
		if strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "/") {
			// absolute/relative filename -> local file (this if block allows copying of files with colons in them)
			continue
		} else if parts := strings.SplitN(arg, ":", 2); len(parts) == 2 {
			// remote file -> parse and transform user@host part
			// results in "[user@]account.devId"
			var tgtHost, tgtUser = sshParseTarget(parts[0])
			if tgtUser != "" {
				tgtHost = fmt.Sprintf("%s@%s", tgtUser, tgtHost)
			}
			args[i] = fmt.Sprintf("%s:%s", tgtHost, parts[1])
		}
	}

	// TODO this will fail if argv[0] contains spaces
	a := []string{scpPath, "-3", fmt.Sprintf("-oProxyCommand=%s pipe %%h ssh", os.Args[0])}
	if sshGetConfig(opts, "UserKnownHostsFile") == "" {
		a = append(a, fmt.Sprintf("-oUserKnownHostsFile=%s", config.GetConfigPath("known_hosts")))
	}

	a = append(a, opts...)
	a = append(a, args...)

	err := syscall.Exec(scpPath, a, os.Environ())
	if err != nil {
		logg.Fatal("Failed to run ", scpPath, ": ", err)
	}

	logg.Fatal("We shouldn't be here")
	return -1
}

// SCPCommand -- implements `ondevice scp`
var SCPCommand = BaseCommand{
	Arguments: "[scp args...] [[user1@]host1:]file1 ... [[userN@]hostN:]fileN",
	ShortHelp: "Copy files from/to your devices using scp",
	RunFn:     scpRun,
	LongHelp: `$ ondevice scp [scp-options...]

Copy files from/to devices using scp

Examples:

- ondevice scp -pv /source/path/ root@myDev:/target/path/
    copy the local /src/path to myDev's /target/path/ as root
    (and pass the -p and -v options to scp)
- ondevice scp me@otherDev:/etc/motd /tmp/other.motd
    copy otherDev's /etc/motd file to /tmp/other.motd (and login as 'me')

Notes:
- while it is possible to copy files between two servers, scp will initiate
  both connections simultaneously causing two concurrent password prompts
	which won't work (unless of course you've set up ssh_agent properly).
- uses scp's '-3' flag (allowing files to be copied between two remote devices)
- We use our own known_hosts file (in ~/.config/ondevice/known_hosts).
  Override with ''-oUserKnownHostsFile=...'
`,
}
