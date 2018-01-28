package daemon

import (
	"fmt"
	"os"
	"syscall"

	"github.com/ondevice/ondevice/logg"
)

type lockFile struct {
	Path   string
	fd     int
	closed bool
}

// TryLock -- Try to acquire the daemon's lock file (and write to PID file)
//
// We don't want `ondevice daemon` to be run more than once per user and pc
// since both instances would request the same devID from the server.
// The API server would then believe there was a conflict and after ~5min
// assign a new devID to one of the daemons.
//
// This issue would be repeated (e.g. the next time the system is restarted)
// and therefore cause a lot of garbage data.
//
func (l *lockFile) TryLock() error {
	var err error
	if l.fd, err = syscall.Open(l.Path, os.O_CREATE|os.O_WRONLY, 0644); err != nil {
		logg.Fatalf("Couldn't open '%s' for locking: %s", l.Path, err)
	}
	l.closed = false

	if err = syscall.Flock(l.fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return fmt.Errorf("ondevice daemon seems to be running already (%s)", err)
	}
	logg.Debug("acquired daemon lock file")

	// only do this if we've got the lock
	logg.Debug("Writing to PID file: ", os.Getpid())
	pidstr := fmt.Sprintf("%d\n", os.Getpid())
	syscall.Write(l.fd, []byte(pidstr))

	return nil
}

func (l *lockFile) Unlock() {
	var err error

	if l.closed {
		return
	}

	if err = syscall.Flock(l.fd, syscall.LOCK_UN); err != nil {
		logg.Fatal("Failed to unlock lock file: ", err)
	}
	if err = syscall.Close(l.fd); err != nil {
		logg.Fatal("Failed to close lock file: ", err)
	}
	l.closed = true
}
