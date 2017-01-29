package command

import (
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/control"
	"github.com/ondevice/ondevice/daemon"
	"github.com/ondevice/ondevice/logg"
	"github.com/ondevice/ondevice/tunnel"
)

const _longDaemonHelp = `ondevice daemon [--conf=...] [--pidfile=...] [--sock=...]

Starts the ondevice daemon (the device side of the tunnels).

On debian based systems instead of running 'ondevice daemon' directly, you
should install the ondevice-daemon package instead (which will also take care
of setting up the credentials)

Make sure you run 'ondevice login' and authenticate with your device key first.

Usually you'll only want to have one 'ondevice daemon' instance per device.
If you want to run multiple, you'll have to specify the .conf, .pid and .sock
files manually.
Concurrent daemon instances can't use the same ondevice.conf file!!!

Options:
--conf=/path/to/ondevice.conf
  Path to the ondevice.conf file
  Default: ~/.config/ondevice/ondevice.conf

--pidfile=/path/to/ondevice.pid
  Path to the ondevice.pid file
  Default: ~/.config/ondevice/ondevice.pid

--sock=/path/to/ondevice.sock
  Path to the ondevice.sock file
  Default: ~/.config/ondevice/ondevice.sock
`

// DaemonCommand -- `ondevice daemon` implementation
type DaemonCommand struct {
}

// DaemonOpts -- commandline arguments for `ondevice daemon`
var DaemonOpts struct {
	Configfile string `long:"conf" description:"Path to ondevice.conf (default: ~/.config/ondevice.conf)"`
	Pidfile    string `long:"pidfile" description:"Path to ondevice.pid (default: ~/.config/ondevice.pid)"`
	Socketpath string `long:"sock" description:"Path to ondevice.sock (default: ~/.config/ondevice.sock)"`
}

func (d *DaemonCommand) args() string {
	return "[--conf=ondevice.conf] [--pidfile=ondevice.pid] [--sock=ondevice.sock]"
}

func (d *DaemonCommand) longHelp() string {
	return _longDaemonHelp
}

func (d *DaemonCommand) shortHelp() string {
	return "Run the ondevice device daemon"
}

func (d *DaemonCommand) run(args []string) int {
	_parseArgs(args)

	if !daemon.TryLock() {
		logg.Fatal("Couldn't acquire lock file")
	}

	c := control.StartServer()

	// TODO implement a sane way to stop this infinite loop (at least SIGTERM, SIGINT and maybe a unix socket call)
	retryDelay := 10 * time.Second
	for true {
		d, err := daemon.Connect()
		if err != nil {
			// only abort here if it's an authentication issue
			if _, ok := err.(tunnel.AuthenticationError); ok {
				logg.Fatal(err)
			}

			// sleep for a bit to avoid spamming the servers
			if retryDelay > 120*time.Second {
				retryDelay = 120 * time.Second
			}
			if retryDelay < 10*time.Second {
				retryDelay = 10 * time.Second
			}

			logg.Errorf("device error - retrying in %ds", retryDelay/time.Second)
			time.Sleep(retryDelay)

			retryDelay = time.Duration(float32(retryDelay) * 1.5)
			continue
		}

		c.Daemon = d
		d.Wait()

		// connection was successful -> restart after 10sec
		logg.Warning("lost device connection, reconnecting in 10s")
		retryDelay = 10
		time.Sleep(retryDelay * time.Second)
	}

	return 0
}

func _parseArgs(args []string) {
	opts := DaemonOpts
	if _, err := flags.ParseArgs(&opts, args); err != nil {
		logg.Fatal(err)
	}

	if opts.Configfile != "" {
		config.SetFilePath("ondevice.conf", opts.Configfile)
	}
	if opts.Pidfile != "" {
		config.SetFilePath("ondevice.pid", opts.Pidfile)
	}
	if opts.Socketpath != "" {
		config.SetFilePath("ondevice.sock", opts.Socketpath)
	}
}
