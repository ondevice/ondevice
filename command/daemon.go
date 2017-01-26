package command

import (
	"time"

	"github.com/ondevice/ondevice/daemon"
	"github.com/ondevice/ondevice/logg"
	"github.com/ondevice/ondevice/tunnel"
)

// DaemonCommand -- `ondevice daemon` implementation
type DaemonCommand struct {
}

func (d DaemonCommand) args() string {
	return "[-f]"
}

func (d DaemonCommand) longHelp() string {
	logg.Fatal("Implement me!!!")
	return ""
}

func (d DaemonCommand) shortHelp() string {
	return "Run the ondevice device daemon"
}

// Run -- implements `ondevice daemon`
func (d DaemonCommand) Run(args []string) int {
	if !daemon.TryLock() {
		logg.Fatal("Couldn't acquire lock file")
	}

	// TODO start the unix socket, etc.
	// TODO implement a sane way to stop this infinite loop (at least SIGTERM, SIGINT and maybe a unix socket call)
	retryDelay := 10 * time.Second
	for true {
		d, err := daemon.Connect()
		if err != nil {
			// only abort here if it's an authentication issue
			if _, ok := err.(tunnel.AuthenticationError); ok {
				logg.Fatal(err)
			}

			// sleep for a bit to avoid spamming the servers
			if retryDelay > 120*time.Second {
				retryDelay = 120 * time.Second
			}
			if retryDelay < 10*time.Second {
				retryDelay = 10 * time.Second
			}

			logg.Errorf("device error - retrying in %ds", retryDelay/time.Second)
			time.Sleep(retryDelay)

			retryDelay = time.Duration(float32(retryDelay) * 1.5)
			continue
		}
		d.Wait()

		// connection was successful -> restart after 10sec
		logg.Warning("lost device connection, reconnecting in 10s")
		retryDelay = 10
		time.Sleep(retryDelay * time.Second)
	}

	return 0
}
