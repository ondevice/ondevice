package command

import "log"

// Command - Defines a command
type Command interface {
	args() string
	shortHelp() string
	longHelp() string
	Run(args []string)
}

// TODO find a way to make me const
var _commands = map[string]Command{
	"help":   new(HelpCmd),
	"device": new(DeviceCmd),
	"list":   new(ListCmd),
	"login":  new(LoginCmd),
}

var deviceCmds = []string{}
var clientCmds = []string{"device", "list"}

// Get -- Return specified command (or nil if not found)
func Get(cmd string) Command {
	return _commands[cmd]
}

// Help -- Get help for a specific command, returning
func Help(cmd string) (args string, short string, long string) {
	return "arg1, arg2", "help", "help meeeeeeeeeeee!"
}

// List -- list command names
func List() map[string]Command {
	return _commands
}

// Run -- Run a command with the specified arguments
func Run(cmdName string, args []string) {
	cmd := Get(cmdName)
	if cmd == nil {
		log.Fatal("Command not found:", cmdName)
	}
	cmd.Run(args)
}
