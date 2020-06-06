package daemon

import (
	"errors"
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
	var paths = []string{
		config.MustLoad().GetFilePath(config.PathOndevicePID),
		"/var/run/ondevice/ondevice.pid",
	}
	var file *os.File

	for _, path := range paths {
		if f, err := os.Open(path); err == nil {
			file = f
			break
		}
	}
	if file == nil {
		return -1, errors.New("Couldn't open PID file")
	}
	defer file.Close()

	buff := make([]byte, 100)
	count, err := file.Read(buff)
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
