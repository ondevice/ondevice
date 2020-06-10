package internal

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
)

// ExecExternalCommand -- uses syscall.Exec() to execute the given command
//
// uses exec.LookPath() to find cmd in path
// syscall.Exec() won't return unless there's an error
// and this function logs a fatal error
// -> this function won't return
func ExecExternalCommand(cmd string, args []string) error {
	var err error
	if c, err := exec.LookPath(cmd); err == nil {
		cmd = c
	} else {
		logrus.WithError(err).Fatalf("failed to find command '%s'", cmd)
		return err
	}

	// syscall.Exec will replace this app with the new one (yes, replace it, not just launch)
	// therefore, unless there's an error, this is the last line of code to be executed
	if err = syscall.Exec(cmd, args, os.Environ()); err != nil {
		logrus.WithError(err).Fatalf("failed to run external command: '%s'", cmd)
		return err
	}

	// nothing here should ever be executed
	logrus.Fatalf("this should never happen")
	return nil
}
