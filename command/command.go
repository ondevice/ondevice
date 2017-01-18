package command

import "log"

// Command - Defines a command
type Command interface {
	args() string
	shortHelp() string
	longHelp() string
	Run(args []string) int
}

// TODO find a way to make me const
var _commands = map[string]Command{
	"device": new(DeviceCmd),
	"help":   new(HelpCmd),
	"list":   new(ListCmd),
	"pipe":   new(PipeCmd),
	"rsync":  new(RsyncCommand),
	"setup":  new(SetupCmd),
	"ssh":    new(SSHCommand),
}

var deviceCmds = []string{}
var clientCmds = []string{"device", "list", "pipe", "rsync", "ssh"}
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
		log.Fatal("Command not found:", cmdName)
	}
	return cmd.Run(args)
}
