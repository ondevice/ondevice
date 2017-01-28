package daemon

import (
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/ondevice/ondevice/config"
)

// GetDaemonProcess -- Returns the Process handler for 'ondevice daemon'
func GetDaemonProcess() (*os.Process, error) {
	pid, err := getDaemonPID()
	if err != nil {
		return nil, err
	}
	rc, err := os.FindProcess(pid)
	if err != nil {
		return nil, err
	}

	err = IsRunning(rc)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

// IsRunning -- Returns nil if the process is running or an error otherwise
func IsRunning(p *os.Process) error {
	return p.Signal(syscall.Signal(0))
}

func getDaemonPID() (int, error) {
	path := config.GetConfigPath("ondevice.pid")
	f, err := os.Open(path)
	if err != nil {
		return -1, err
	}

	defer f.Close()

	buff := make([]byte, 100)
	count, err := f.Read(buff)
	if err != nil {
		return -1, err
	}

	s := strings.Trim(string(buff[:count]), " \t\n")
	pid, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return -1, err
	}

	return int(pid), nil
}
