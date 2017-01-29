package command

import (
	"net/url"
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

--sock=unix:///path/to/ondevice.sock
  Path to the ondevice.sock file
  Default: unix://~/.config/ondevice/ondevice.sock


Example Socket URLs:
- unix:///home/user/.config/ondevice/ondevice.sock
  User's ondevice.sock path - clients will use this URL first when connecting
- unix:///var/run/ondevice.sock
  Default system-wide ondevice.sock path - if the above failed, clients will try
  this one instead.
- /var/run/ondevice.sock
  Same as the above (since unix:// is the default URL scheme here)
- http://localhost:1234/
	Listen on TCP port 1234.
  Note that there's currently support for neither SSL nor authentication so use
  this only if absolutely necessary

On the client side, set the ONDEVICE_HOST environment variable to match the
socket parameter.
`

// DaemonCommand -- `ondevice daemon` implementation
type DaemonCommand struct {
}

// DaemonOpts -- commandline arguments for `ondevice daemon`
var DaemonOpts struct {
	Configfile string `long:"conf" description:"Path to ondevice.conf (default: ~/.config/ondevice.conf)"`
	Pidfile    string `long:"pidfile" description:"Path to ondevice.pid (default: ~/.config/ondevice.pid)"`
	SocketURL  string `long:"sock" description:"ondevice.sock URL (default: unix://~/.config/ondevice.sock)"`
}

func (d *DaemonCommand) args() string {
	return "[--conf=ondevice.conf] [--pidfile=ondevice.pid] [--sock=unix://ondevice.sock]"
}

func (d *DaemonCommand) longHelp() string {
	return _longDaemonHelp
}

func (d *DaemonCommand) shortHelp() string {
	return "Run the ondevice device daemon"
}

func (d *DaemonCommand) run(args []string) int {
	url := _parseArgs(args)

	if !daemon.TryLock() {
		logg.Fatal("Couldn't acquire lock file")
	}

	c := control.StartServer(url)

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

func _parseArgs(args []string) url.URL {
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
	if opts.SocketURL != "" {
		if rc, err := url.Parse(opts.SocketURL); err != nil {
			logg.Fatal("Couldn't parse socket URL: ", err)
		} else {
			return *rc
		}
	}

	return url.URL{Scheme: "unix", Path: config.GetConfigPath("ondevice.sock")}
}
