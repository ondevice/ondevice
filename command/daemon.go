package command

import (
	"fmt"
	"log"
	"net/url"
	"os"

	flags "github.com/jessevdk/go-flags"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/control"
	"github.com/ondevice/ondevice/daemon"
	"github.com/ondevice/ondevice/logg"
)

// DaemonOpts -- commandline arguments for `ondevice daemon`
type DaemonOpts struct {
	Configfile string `long:"conf" description:"Path to ondevice.conf (default: ~/.config/ondevice.conf)"`
	Pidfile    string `long:"pidfile" description:"Path to ondevice.pid (default: ~/.config/ondevice.pid)"`
	SocketURL  string `long:"sock" description:"ondevice.sock URL (default: unix://~/.config/ondevice.sock)"`
}

func daemonRun(args []string) int {
	// main.go disables timestamps in log messages. re-enable them
	log.SetFlags(log.LstdFlags)

	if os.Getuid() == 0 {
		logg.Fatal("`ondevice daemon` should not be run as root")
	}

	var d = daemon.NewDaemon()
	var controlURL url.URL
	var err error

	if controlURL, err = daemonParseArgs(args, d); err != nil {
		logg.Fatal(err)
		return 1
	}

	c := control.NewSocket(d, controlURL)
	d.Control = c

	return d.Run()
}

// Parses the commandline arguments, returns the ControlSocket URL
func daemonParseArgs(args []string, d *daemon.Daemon) (url.URL, error) {
	var opts DaemonOpts
	var rc url.URL
	var err error

	if args, err = flags.ParseArgs(&opts, args); err != nil {
		return rc, err
	}

	if len(args) > 0 {
		return rc, fmt.Errorf("Too many arguments: %s", args)
	}

	d.PIDFile = opts.Pidfile

	if opts.Configfile != "" {
		config.SetFilePath("ondevice.conf", opts.Configfile)
	}

	if opts.Pidfile == "" {
		d.PIDFile = config.GetConfigPath("ondevice.pid")
	}

	if opts.SocketURL != "" {
		if rc, err := url.Parse(opts.SocketURL); err != nil {
			logg.Fatal("Couldn't parse socket URL: ", err)
		} else {
			return *rc, nil
		}
	}
	return url.URL{Scheme: "unix", Path: config.GetConfigPath("ondevice.sock")}, nil
}

// DaemonCommand -- implements `ondevice daemon`
var DaemonCommand = BaseCommand{
	Arguments: "[--conf=ondevice.conf] [--pidfile=ondevice.pid] [--sock=unix://ondevice.sock]",
	ShortHelp: "Run the ondevice device daemon",
	RunFn:     daemonRun,
	LongHelp: `$ ondevice daemon [--conf=...] [--pidfile=...] [--sock=...]

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
- unix:///var/run/ondevice/ondevice.sock
  Default system-wide ondevice.sock path - if the above failed, clients will try
  this one instead.
- /var/run/ondevice/ondevice.sock
  Same as the above (since unix:// is the default URL scheme here)
- http://localhost:1234/
	Listen on TCP port 1234.
  Note that there's currently support for neither SSL nor authentication so use
  this only if absolutely necessary

On the client side, set the ONDEVICE_HOST environment variable to match the
socket parameter.
`,
}
