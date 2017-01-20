package daemon

import (
	"fmt"
	"log"
	"os"
	"syscall"

	"github.com/ondevice/ondevice/config"
)

// TryLock -- Try to acquire the daemon's lock file (and write to PID file)
func TryLock() bool {

	lockFile := config.GetConfigPath("ondevice.lock")

	fd, err := syscall.Open(lockFile, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		log.Fatal("Couldn't open lock file: ", err)
	}

	if err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		log.Fatalf("ondevice daemon seems to be running already (%s)", err)
	}
	log.Print("Ackquired daemon lock file")

	// only do this if we've got the
	log.Print("Writing to PID file: ", os.Getpid())
	pidFile, err := os.Create(config.GetConfigPath("ondevice.pid"))
	if err != nil {
		log.Fatal("Couldn't open PID file: ", err)
	}
	pidFile.WriteString(fmt.Sprintf("%d\n", os.Getpid()))

	return true
}
