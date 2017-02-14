### v0.4.3 (2017-02-14)

Bugfix release:

- added support for the `ONDEVICE_LOG` environment variable to set the log level
- fixed issue with closing tunnel sessions and slightly cleaned up the code responsible for it

### v0.4.2 (2017-02-04)

Bugfix release:

- fixed issue with cleanly closing connections
- printing byte counts when connections close
- improved certain error messages
- some other minor improvements

### v0.4.1 (2017-01-30)

Bugfix release:

- fixed an issue that prevented the environment to be preserved when running `ondevice ssh` or `ondevice rsync`
- allowing root to run everything but `ondevice daemon`
- there are some minor packaging issues in [ondevice-debian][ondevice-debian] that have been addressed

### v0.4.0

- initial golang-based release
