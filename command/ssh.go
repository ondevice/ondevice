package command

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

const sshFlags = "1246ab:c:e:fgi:kl:m:no:p:qstvxD:L:NR:"

const _longSSHHelp = `
Connect to your devices using the 'ssh' command.

Usage:
    ondevice ssh [<user>@]<device> [ssh-arguments...]

This is a relatively thin wrapper around the 'ssh' command.
The main difference to invoking ssh directly is that instead of regular host names you'll have to specify an ondevice deviceId.
The connection is routed through the ondevice.io network.

ondevice ssh will try to parse ssh's arguments, the first non-argument has to be
the user@hostname combo.

See ssh's documentation for further details.

Examples:
- ondevice ssh device1
    simply connect to device1
- ondevice ssh user@device1
    open an SSH connection to device1, logging in as 'user'
- ondevice ssh device1 echo hello world
    run 'echo hello world' on device1
- ondevice ssh device1 -N -L 1234:localhost:80
    Tunnel the HTTP server on device1 to the local port 1234 without opening
    a shell
- ondevice ssh device1 -D 1080
    Starting a SOCKS5 proxy listening on port 1080. It'll redirect all traffic
    to the target host.
`

// SSHCommand -- implements `ondevice ssh`
type SSHCommand struct{}

func (s SSHCommand) args() string {
	return "[ssh-arguments...]"
}

func (s SSHCommand) longHelp() string {
	return _longSSHHelp
}

func (s SSHCommand) shortHelp() string {
	return "Connect to your devices using the ssh protocol"
}

func (s SSHCommand) run(args []string) int {
	sshPath := "/usr/bin/ssh"

	// parse args (to detect the ones before 'user@host')
	target, args := s._parseArgs(args)
	tgtHost, tgtUser := s._parseTarget(target)

	// compose ProxyCommand
	// TODO this will fail miserably if argv[0] or tgtHost contain spaces
	proxyCmd := fmt.Sprintf("-oProxyCommand=%s pipe %s ssh", os.Args[0], tgtHost)

	// create something like `ssh -oProxyCommand=... user@ondevice:devId <opts`
	a := make([]string, 0, 10)
	a = append(a, sshPath, proxyCmd)
	if tgtUser != "" {
		a = append(a, fmt.Sprintf("%s@ondevice:%s", tgtUser, tgtHost))
	} else {
		a = append(a, fmt.Sprintf("ondevice:%s", tgtHost))
	}
	a = append(a, args...)

	// syscall.Exec will replace this app with ssh (yes, replace it, not just launch)
	// therefore, unless there's an error, this is the last line of code to be executed
	err := syscall.Exec(sshPath, a, os.Environ())
	if err != nil {
		logg.Fatal("Failed to run ", sshPath, ": ", err)
	}

	// nothing here should ever be executed
	logg.Fatal("This should never happen")
	return -1
}

func (s SSHCommand) _parseArgs(args []string) (string, []string) {
	// option list copied from debian jessie's openssh source package (from ssh.c, line 509)
	flags := _getSSHFlags()
	var target string
	var outArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]

		if len(arg) == 0 {
			logg.Fatal("Failed to parse SSH arguments: got an empty argument while looking for '[user@]host'")
		} else if arg == "-" {
			logg.Fatal("Stray '-' SSH argument")
		} else if arg[0] == '-' {
			hasValue, ok := flags[arg[1]]
			if !ok {
				logg.Fatal("Unsupported SSH argument: ", arg)
			}
			if hasValue && len(arg) == 2 {
				// the value's in the next argument, push both to outArgs
				outArgs = append(outArgs, arg, args[i+1])
				i++
			} else if hasValue && len(arg) > 2 {
				// the value's part of arg
				outArgs = append(outArgs, arg)
			} else if !hasValue && len(arg) > 2 { // && !hasValue
				logg.Fatal("Got value for flag that doesn't expect one: ", arg)
			} else if !hasValue && len(arg) == 2 {
				// a simple flag)
				outArgs = append(outArgs, arg)
			} else {
				// yay to defensive programming
				logg.Fatal("this should never happen (fifth state of two binary values)")
			}
		} else {
			// first non-option argument -> extract target and keep the rest as-is
			target = arg
			outArgs = append(outArgs, args[i+1:]...)
			return target, outArgs
		}
	}

	logg.Fatal("Missing SSH target user/host!")
	return "", nil
}

func (s SSHCommand) _parseTarget(target string) (tgtHost string, tgtUser string) {
	parts := strings.SplitN(target, "@", 2)
	if len(parts) == 1 {
		tgtUser = ""
		tgtHost = parts[0]
	} else if len(parts) == 2 {
		tgtUser = parts[0]
		tgtHost = parts[1]
	} else {
		logg.Fatal("SplitN(..., 2) returned odd number of results: ", len(parts))
	}

	// always use qualified device names
	if !strings.Contains(tgtHost, ".") {
		user, _, err := config.GetClientAuth()
		if err == nil {
			tgtHost = fmt.Sprintf("%s.%s", user, tgtHost)
		}
	}

	return tgtHost, tgtUser
}

func _getSSHFlags() map[byte]bool {
	rc := map[byte]bool{}

	for i := 0; i < len(sshFlags); i++ {
		flag := sshFlags[i]
		hasValue := false
		if sshFlags[i+1] == ':' {
			hasValue = true
			i++
		}
		rc[flag] = hasValue
	}
	return rc
}
