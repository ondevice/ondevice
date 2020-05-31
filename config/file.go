package config

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

// Read -- fetches the contents of ondevice.conf
func Read() (Config, error) {
	var rc Config
	var err error

	rc.path = GetConfigPath("ondevice.conf")
	if rc.cfg, err = ini.InsensitiveLoad(rc.path); err != nil {
		if !os.IsNotExist(err) {
			logrus.WithError(err).Error("config.Read(): failed to read ondevice.conf")
		}
		rc.cfg = ini.Empty()
		return rc, err
	}
	return rc, nil
}

// writeFile -- writes the given data to targetPath
//
// Meticulous care needs to be taken not to end up with corrupted files
// (worst case: we lose authentication info and a device becomes unreachable)
// We therefore write to a temporary file in the same directory first and overwrite 'path' only if things were successful
//
// things to keep in mind
// - only call this function with VALID configuration - i.e. ones you've tried out
// - we might not have permission creating the temporary file
// - we might not have permissions to replace the original file
// - disk space may have run out
// - one or more other instances may compete with us - in which case one should succeed and the others may even fail silently
//   to ensure that, only os.Rename() actually competes with other instances -> the last one calling it will win.
//   potential damage caused by this is small
// - ONLY IF everything succeeds will we invoke os.Rename()
//
// Note that we rely on the following OS-level behaviour:
// - ioutil.TempFile() returns a file (and path) that is truly ours and won't be modified by anyone else while we're using it
// - chmod operates on the file descriptor instead of the path (i.e. actually changes the mode of the file we're writing to)
// - after File.Write() (or at least after File.Close()) the contents of the file are accessible by others (i.e. have been flushed to the filesystem layer)
// - os.Rename() is atomic (i.e. updates the inode of the target file before unlinking the source)
// we might need to add special implementations for Windows and other systems
func writeFile(data []byte, path string, filemode os.FileMode) error {
	var file *os.File
	var err error
	var dir = filepath.Dir(path)

	// TODO we may need an os.Makedirs() here

	if file, err = ioutil.TempFile(dir, ".tmp_"); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to create temporary config file")
		return err
	}

	// make sure the file it gets closed and deleted no matter what happens (will of course leak temporary files if terminated forcibly)
	// these two calls are expected to fail -> ignore errors
	var tmpPath = file.Name()
	defer os.Remove(tmpPath)
	defer file.Close()

	if err = file.Chmod(filemode); err != nil {
		logrus.WithError(err).WithField("path", path).Error("failed to chmod temporary config file")
		return err
	}

	var n int
	// according to the documentation, Write() doesn't buffer things
	if n, err = file.Write(data); err != nil {
		logrus.WithError(err).WithField("path", tmpPath).Error("failed to write to config file")
		return err
	} else if n != len(data) {
		logrus.WithField("path", tmpPath).Errorf("got odd number of bytes while writing to config file: got=%d, expected=%d", n, len(data))
		return errors.New("not all data could be written to the temporary config file")
	}

	// just to make sure the data is flushed to the filesystem layer
	if err = file.Close(); err != nil {
		logrus.WithError(err).WithField("path", tmpPath).Error("failed to close temporary config file - this should not happen")
		return err
	}

	if err = os.Rename(tmpPath, path); err != nil {
		logrus.WithError(err).WithField("tmpPath", tmpPath).WithField("path", path).Errorf("failed to overwrite '%s' - make sure to check file/directory permissions", filepath.Base(path))
		return err
	}

	return nil
}
