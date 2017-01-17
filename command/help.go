package command

import (
	"fmt"
	"log"
	"os"
)

const _longHelpHelp = `ondevice help [cmd]

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

func (h HelpCommand) args() string {
	return "[cmd]"
}

func (h HelpCommand) longHelp() string {
	return _longHelpHelp
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
	l := log.New(os.Stderr, "", 0)
	l.Println("USAGE: ondevice <command> [...]")

	cmds := List()

	l.Println("\n- Device commands:")
	h._listCommands(deviceCmds, cmds)

	l.Println("\n- Client commands:")
	h._listCommands(clientCmds, cmds)

	l.Println("\n- Other commands:")
	h._listCommands(nil, cmds)
}

func (h HelpCommand) _listCommands(names []string, cmds map[string]Command) {
	if names == nil {
		names = []string{}
		for k := range cmds {
			names = append(names, k)
		}
	}

	for i := range names {
		name := names[i]
		if _, ok := cmds[name]; !ok {
			log.Fatal("Command not found: ", name)
		}
		cmd := cmds[name]
		fmt.Printf("    %s %s\n", name, cmd.args())
		fmt.Println("        ", cmd.shortHelp())

		delete(cmds, name)
	}
}

func (h HelpCommand) help(cmdName string) {
	cmd := Get(cmdName)
	if cmd == nil {
		log.Fatal("Command not found:" + cmdName)
	} else {
		fmt.Println(cmd.longHelp())
	}
}
