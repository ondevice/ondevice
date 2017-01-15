package command

import (
	"fmt"
	"log"
)

var _longHelp = `ondevice help [cmd]

Lists commands (if [cmd] was omitted) or shows details for a specific command)

Examples:
    - ondevice help
      lists available commands
    - ondevice help help
      shows help for the 'ondevice help' command
    - ondevice help login
      shows help for the 'ondevice login' command
`

// HelpCommand - the 'help' command
type HelpCommand struct {
}

func (h HelpCommand) args() []string {
	return nil
}

func (h HelpCommand) longHelp() string {
	return _longHelp
}

func (h HelpCommand) shortHelp() string {
	return "Shows this help screen"
}

// Run -- run `ondevice help <args>`
func (h HelpCommand) Run(args []string) {
	if len(args) == 0 {
		h.listCommands()
	} else if len(args) == 1 {
		cmd := args[0]
		h.help(cmd)
	} else {
		log.Fatal("USAGE: ondevice help [cmd]")
	}
}

// ListCommands -- implements `ondevice help`
func (h HelpCommand) listCommands() {
	cmds := List()
	for i := range cmds {
		cmd := Get(cmds[i])
		fmt.Println(cmds[i], cmd.shortHelp())
	}
}

func (h HelpCommand) help(cmdName string) {
	cmd := Get(cmdName)
	if cmd == nil {
		log.Fatal("Command not found:" + cmdName)
	} else {
		fmt.Println("help " + cmdName + ": " + cmd.longHelp())
	}
}
