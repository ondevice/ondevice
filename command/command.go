package command

import "github.com/ondevice/ondevice/logg"

// Command - Defines a command
type Command interface {
	args() string
	shortHelp() string
	longHelp() string
	run(args []string) int
}

// TODO find a way to make me const
var _commands = map[string]Command{
	"daemon": new(DaemonCommand),
	"device": new(DeviceCmd),
	"help":   new(HelpCmd),
	"list":   new(ListCmd),
	"login":  new(LoginCmd),
	"rsync":  new(RsyncCommand),
	"ssh":    new(SSHCommand),
	"status": new(StatusCmd),
	"stop":   new(StopCmd),
	"pipe":   new(PipeCmd),
}

var deviceCmds = []string{"daemon", "stop"}
var clientCmds = []string{"device", "list", "pipe", "rsync", "ssh"}

// hide these commands from `ondevice help`
var internalCmds = map[string]int{"pipe": 0}

// Get -- Return specified command (or nil if not found)
func Get(cmd string) Command {
	return _commands[cmd]
}

// List -- list command names
func List() map[string]Command {
	return _commands
}

// Run -- Run a command with the specified arguments
func Run(cmdName string, args []string) int {
	cmd := Get(cmdName)
	if cmd == nil {
		logg.Fatal("Command not found:", cmdName)
	}
	return cmd.run(args)
}
