package command

import (
	"fmt"
	"log"
	"os"

	"github.com/ondevice/ondevice/logg"
)

func helpRun(args []string) int {
	if len(args) == 0 {
		helpListCommands()
	} else if len(args) == 1 {
		cmd := args[0]
		helpCommand(cmd)
	} else {
		logg.Fatal("USAGE: ondevice help [cmd]")
	}

	return 0
}

// helpListCommands -- implements `ondevice help` (without arguments)
func helpListCommands() {
	l := log.New(os.Stderr, "", 0)
	l.Println("USAGE: ondevice <command> [...]")

	cmds := List()
	showInternal := false // TODO use a commandline flag or something

	l.Println("\n- Device commands:")
	helpListCommandsByName(deviceCmds, cmds, showInternal)

	l.Println("\n- Client commands:")
	helpListCommandsByName(clientCmds, cmds, showInternal)

	l.Println("\n- Other commands:")
	helpListCommandsByName(nil, cmds, showInternal)

	l.Println()
}

func helpListCommandsByName(names []string, cmds map[string]Command, showInternal bool) {
	if names == nil {
		names = []string{}
		for k := range cmds {
			names = append(names, k)
		}
	}

	for _, name := range names {
		if _, ok := internalCmds[name]; !showInternal && ok {
			continue // skip internal commands (unless showInternal is true)
		}

		if _, ok := cmds[name]; !ok {
			logg.Fatal("Command not found: ", name)
		}
		cmd := cmds[name]
		fmt.Printf("    %s %s\n", name, cmd.args())
		fmt.Println("        ", cmd.shortHelp())

		delete(cmds, name)
	}
}

func helpCommand(cmdName string) {
	cmd := Get(cmdName)
	if cmd == nil {
		logg.Fatal("Command not found: " + cmdName)
	} else {
		fmt.Println(cmd.longHelp())
	}
}

func init() {
	// Registering the `help` command dynamically here to avoid initialization loops
	Register("help", BaseCommand{
		Arguments: "[cmd]",
		ShortHelp: "Shows this help screen",
		RunFn:     helpRun,
		LongHelp: `$ ondevice help [cmd]

Lists commands (if [cmd] was omitted) or shows details for a specific command)

Examples:
    - ondevice help
      lists available commands
    - ondevice help help
      shows help for the 'ondevice help' command
    - ondevice help login
      shows help for the 'ondevice login' command
`,
	})
}
