package command

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ondevice/ondevice/logg"
)

func rsyncRun(args []string) int {
	rsyncPath := "/usr/bin/rsync"

	// TODO this will fail if argv[0] contains spaces
	a := []string{rsyncPath, "-e", fmt.Sprintf("%s ssh", os.Args[0])}
	a = append(a, args...)

	err := syscall.Exec(rsyncPath, a, os.Environ())
	if err != nil {
		logg.Fatal("Failed to run ", rsyncPath, ": ", err)
	}

	logg.Fatal("We shouldn't be here")
	return -1
}

// RsyncCommand -- implements `ondevice rsync`
var RsyncCommand = BaseCommand{
	Arguments: "[rsync args...]",
	ShortHelp: "Copy files from/to your devices using rsync",
	RunFn:     rsyncRun,
	LongHelp: `$ ondevice rsync [rsync-options...]

Copy files from/to devices using rsync

Examples:

- ondevice rsync -av /source/path/ root@myDev:/target/path/
    copy the local /src/path to myDev's /target/path/ as root
    (and pass the -a and -v options to rsync)
- ondevice rsync me@otherDev:/etc/motd /tmp/other.motd
    copy otherDev's /etc/motd file to /tmp/other.motd (and login as 'me')

This command is only a thin wrapper around the 'rsync' client (using its '-e'
argument to make it use 'ondevice ssh' internally).
`,
}
