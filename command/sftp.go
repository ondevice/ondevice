package command

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

var sftpFlags = sshParseFlags("1246aCfpqrvB:b:c:D:F:i:l:o:P:R:S:s:")

func runSFTP(args []string) int {
	sftpPath := "/usr/bin/sftp"

	args, opts := sshParseArgs(sftpFlags, args)

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
	a := []string{sftpPath, fmt.Sprintf("-oProxyCommand=%s pipe %%h ssh", os.Args[0])}
	if sshGetConfig(opts, "UserKnownHostsFile") == "" {
		a = append(a, fmt.Sprintf("-oUserKnownHostsFile=%s", config.GetConfigPath("known_hosts")))
	}

	a = append(a, opts...)
	a = append(a, args...)

	err := syscall.Exec(sftpPath, a, os.Environ())
	if err != nil {
		logg.Fatal("Failed to run ", sftpPath, ": ", err)
	}

	logg.Fatal("We shouldn't be here")
	return -1
}

// SFTPCommand -- implements `ondevice sftp`
var SFTPCommand = BaseCommand{
	Arguments: "[sftp-flags] [user@]devId",
	ShortHelp: "copy files from/to a device using sftp",
	RunFn:     runSFTP,
	LongHelp: `
  $ ondevice sftp [sftp-options...] [user@]devId

  Interactively copy files from/to devices using scp

  Examples:
  - ondevice sftp user@myDev
    open an sftp session to 'myDev', logging in as 'user'

  Notes:
  - We use our own known_hosts file (in ~/.config/ondevice/known_hosts).
  Override with ''-oUserKnownHostsFile=...'
`,
}
