### v0.5.3 (2018-01-??)

- added ondevice list filter expressions (see https://docs.ondevice.io/commands/list/ )
- you can now delete devices using `ondevice device $devId rm on:id`
- added systemd support
- setting request timeouts (preventing the tool to wait indefinitely)
- improved support for AuthKey roles
- improved error handline and logging
- improved build pipeline (automatically rsyncing pushes to https://repo.ondevice.io/builds/ )
- making it clear that you can't use your account password (for `ondevice login` and `dpkg-reconfigure ondevice-dameon`)
- explicitly handling HTTP 429 Too Many Requests when reconnecting in `ondevice daemon`


### v0.5.2 (2017-11-27)

- Added `ondevice scp` and `ondevice sftp`
- using our own known_hosts file (since we're not using domain names, but devIds)
- (using '%h' for ProxyCommand)
- Note: incorrectly identifies as v0.5.1 (sry for that :( )

### v0.5.1 (2017-11-09)

Bugfix release:

- Fixed issue in `ondevice login`
- added `--batch=username` flag to `ondevice login`
- fixed issue with `ondevice list` on 32bit systems (were using a 32bit integer for the timestamp)

### v0.5 (2017-11-01)

- added `ondevice event`
- added support for an official [ondevice/ondevice](https://hub.docker.com/r/ondevice/ondevice) docker image
- added tunnel state machine (to fix state transition issues)
- updated `ondevice login` to support the new AuthKey permission model
- refactored the way commands are implemented (improving code readability)
- added nicer error messages for 'authentication failed' and 'rate limit exceeded'
- doing a os.Chmod(0600) when writing ondevice.conf 
- a lot of minor fixes and improvements

### v0.4.4 (2017-04-02)

Bugfix release:

- using [glide](https://glide.sh/) for dependency management
- added timeout to OpenWebsocket() (fixes rare reconnect issue, see #14)
- Config.SetValue() now recursively creates the config dir (if it doesn't exist)
- fixed issue with ondevice ssh not properly exiting when the connection was lost (see #15)
- fixed synchronisation issue in websocket.Write()

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
