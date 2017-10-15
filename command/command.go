package command

import "github.com/ondevice/ondevice/logg"

// Command - Defines a command
type Command interface {
	args() string
	shortHelp() string
	longHelp() string
	run(args []string) int
}

// BaseCommand -- implements the Command interface
type BaseCommand struct {
	Arguments string
	ShortHelp string
	LongHelp  string
	RunFn     func(args []string) int
}

func (c BaseCommand) args() string {
	return c.Arguments
}

func (c BaseCommand) shortHelp() string {
	return c.ShortHelp
}

func (c BaseCommand) longHelp() string {
	return c.LongHelp
}

func (c BaseCommand) run(args []string) int {
	return c.RunFn(args)
}

var _commands = map[string]Command{
	"daemon": DaemonCommand,
	"device": DeviceCommand,
	"event":  EventCommand,
	"list":   ListCommand,
	"login":  LoginCommand,
	"rsync":  RsyncCommand,
	"ssh":    SSHCommand,
	"status": StatusCommand,
	"stop":   StopCommand,
	"pipe":   PipeCommand,
}

var deviceCmds = []string{"daemon", "stop"}
var clientCmds = []string{"device", "event", "list", "pipe", "rsync", "ssh"}

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
		logg.Fatal("Command not found: ", cmdName)
	}
	return cmd.run(args)
}

// Register -- Registers a new (or replaces an existing) command
//
// pass cmd=nil to remove a Command
func Register(cmdName string, cmd Command) {
	if cmd != nil {
		_commands[cmdName] = cmd
	} else {
		delete(_commands, cmdName)
	}

}
