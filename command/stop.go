package command

import (
	"syscall"
	"time"

	"github.com/ondevice/ondevice/daemon"
	"github.com/ondevice/ondevice/logg"
)

// StopCmd -- Stops the ondevice daemon (by sending SIGTERM)
type StopCmd struct{}

const _longStopHelp = `ondevice stop

Finds a running ondevice daemon (using the ondevice.pid file) and

`

func (s *StopCmd) args() string {
	return ""
}

func (s *StopCmd) longHelp() string {
	return ""
}

func (s *StopCmd) shortHelp() string {
	return "Stops the local ondevice daemon (if running)"
}

func (s *StopCmd) run(args []string) int {
	p, err := daemon.GetDaemonProcess()
	if err != nil {
		logg.Debug("GetDaemonProcess error: ", err)
		logg.Fatal("Couldn't find daemon process")
	}
	logg.Infof("Stopping ondevice daemon... (pid: %d)", p.Pid)

	found := false
	for i := 0; i < 5; i++ {
		if daemon.IsRunning(p) != nil {
			if found {
				// we seem to have stopped it
				return 0
			}
			logg.Debug(err)
			logg.Info("Not running")
			return 1
		}
		found = true
		p.Signal(syscall.SIGTERM)
		time.Sleep(1000 * time.Millisecond)
	}

	logg.Fatal("Failed to stop ondevice daemon in time")
	return 1
}
