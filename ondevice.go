package main

import (
	"os"
	"os/user"
	"strconv"
	"syscall"

	"github.com/ondevice/ondevice/command"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

func main() {
	if os.Getuid() == 0 {
		// running as root -> setup the files we need and drop privileges
		uid := _dropPrivileges()
		if err := syscall.Setuid(int(uid)); err != nil {
			logg.Fatal("Failed to drop privileges: ", err)
		}
	}

	if len(os.Args) < 2 {
		logg.Fatal("Missing command! try 'ondevice help'")
	}

	//logg.Debug("-- args: ", os.Args[1:])
	cmd := os.Args[1]
	rc := command.Run(cmd, os.Args[2:])
	os.Exit(rc)
}

func _dropPrivileges() int {
	// first see if there's an 'ondevice' user account
	u, err := user.Lookup("ondevice")
	if err != nil {
		logg.Fatal("Can't run as root - and couldn't find 'ondevice' user")
	}

	// get uid
	uid, err := strconv.ParseInt(u.Uid, 10, 32)
	if err != nil {
		logg.Fatal("Couldn't convert uid string: ", u.Uid)
	}

	gid, err := strconv.ParseInt(u.Gid, 10, 32)
	if err != nil {
		logg.Fatal("Couldn't convert gid string: ", u.Gid)
	}

	// see if ondevice.conf exists
	_, err = os.Stat("/etc/ondevice/ondevice.conf")
	if os.IsNotExist(err) {
		logg.Fatal("Couldn't find /etc/ondevice/ondevice.conf")
	}

	// TODO use other paths for other OSs
	// TODO allow the user to override these paths (e.g. using environment vars or commandline flags)
	_setupFile("ondevice.pid", "/var/run/ondevice.pid", int(uid), int(gid), 0644)
	_setupFile("ondevice.sock", "/var/run/ondevice.sock", int(uid), int(gid), 0664)
	config.SetFilePath("ondevice.conf", "/etc/ondevice/ondevice.conf")

	return int(uid)
}

func _setupFile(filename string, path string, uid int, gid int, mode os.FileMode) {
	config.SetFilePath(filename, path)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		os.OpenFile(path, os.O_RDONLY|os.O_CREATE, mode)
	} else if err != nil {
		logg.Fatal("Couldn't get file info for ", filename, ": ", err)
	}
	err = os.Chown(path, uid, gid)
	if err != nil {
		logg.Fatalf("Couldn't set file permissions for %s to %d", filename, mode)
	}
}
