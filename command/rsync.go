package command

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ondevice/ondevice/logg"
)

const _longRsyncHelp = `Copy files from/to devices using rsync

Usage:

ondevice rsync [rsync-options...]

Examples:

- ondevice rsync -av /source/path/ root@myDev:/target/path/
    copy the local /src/path to myDev's /target/path/ as root
    (and pass the -a and -v options to rsync)
- ondevice rsync me@otherDev:/etc/motd /tmp/other.motd
    copy otherDev's /etc/motd file to /tmp/other.motd (and login as 'me')

This command is only a thin wrapper around the 'rsync' client (using its '-e'
argument to make it use 'ondevice ssh' internally).
`

// RsyncCommand -- implements `ondevice rsync`
type RsyncCommand struct{}

func (r RsyncCommand) args() string {
	return "[rsync args...]"
}

func (r RsyncCommand) shortHelp() string {
	return "Copy files from/to your devices using rsync"
}

func (r RsyncCommand) longHelp() string {
	return _longRsyncHelp
}

// Run -- `ondevice rsync` implementation
func (r RsyncCommand) Run(args []string) int {
	rsyncPath := "/usr/bin/rsync"

	// TODO this will fail if argv[0] contains spaces
	a := []string{rsyncPath, "-e", fmt.Sprintf("%s ssh", os.Args[0])}
	a = append(a, args...)

	err := syscall.Exec(rsyncPath, a, nil)
	if err != nil {
		logg.Fatal("Failed to run ", rsyncPath, ": ", err)
	}

	logg.Fatal("We shouldn't be here")
	return -1
}
