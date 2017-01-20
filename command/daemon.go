package command

import (
	"log"

	"github.com/ondevice/ondevice/daemon"
)

// DaemonCommand -- `ondevice daemon` implementation
type DaemonCommand struct {
}

func (d DaemonCommand) args() string {
	return "[-f]"
}

func (d DaemonCommand) longHelp() string {
	log.Fatal("Implement me!!!")
	return ""
}

func (d DaemonCommand) shortHelp() string {
	return "Run the ondevice device daemon"
}

// Run -- implements `ondevice daemon`
func (d DaemonCommand) Run(args []string) int {
	if !daemon.TryLock() {
		log.Fatal("Couldn't acquire lock file")
	}

	// TODO start the unix socket, etc.
	// TODO implement a sane way to stop this infinite loop (at least SIGTERM, SIGINT and maybe a unix socket call)
	for true {
		d, err := daemon.Connect()
		if err != nil {
			// TODO only abort here if it's an authentication issue
			log.Fatal(err)
			// TODO sleep for a bit to avoid spamming the server
		}
		d.Wait()
	}

	return 0
}
