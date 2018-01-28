package command

import (
	"syscall"
	"time"

	"github.com/ondevice/ondevice/daemon"
	"github.com/ondevice/ondevice/logg"
)

func runStop(args []string) int {
	var found = false
	for i := 0; i < 5; i++ {
		// fetch the PID every time (the daemon might go inactive - still running until the remaining tunnels close)
		p, err := daemon.GetDaemonProcess()
		if err != nil {
			if !found {
				logg.Debug("GetDaemonProcess error: ", err)
				logg.Fatal("Couldn't find daemon process")
				return 1
			}
			// we seem to have stopped it
			return 0
		}

		if !found {
			logg.Infof("Stopping ondevice daemon... (pid: %d)", p.Pid)
			found = true
		}

		p.Signal(syscall.SIGTERM)
		time.Sleep(1000 * time.Millisecond)
	}

	logg.Fatal("Timeout trying to stop the ondevice daemon")
	return 1
}

// StopCommand -- Stops the ondevice daemon (by sending SIGTERM)
var StopCommand = BaseCommand{
	Arguments: "",
	LongHelp: `$ ondevice stop

Stops a running ondevice daemon (using the ondevice.pid file) and tries to terminate it.

Returns 0 on success or 1 on error
`,
	ShortHelp: "Stops the local ondevice daemon (if running)",
	RunFn:     runStop,
}
