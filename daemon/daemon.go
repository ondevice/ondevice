package daemon

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ondevice/ondevice/logg"
	"github.com/ondevice/ondevice/util"
)

// ControlSocket -- REST API to control the ondevice daemon (the implementation's in the control package)
type ControlSocket interface {
	Start()
	Stop() error
}

// Daemon -- represents a device's connection to the ondevice.io API server
type Daemon struct {
	PIDFile string

	ws            *deviceSocket
	signalChan    chan os.Signal
	firstSIGTERM  time.Time
	lock          lockFile
	activeTunnels sync.WaitGroup
	shutdown      bool

	Control      ControlSocket
	OnConnection func(tunnelID string, service string, protocol string)
	OnError      func(error)
}

type pingMsg struct {
	Type string `json:"_type"`
	Ts   int    `json:"ts"`
}

// NewDaemon -- Create a new Daemon instance
func NewDaemon() *Daemon {
	return &Daemon{
		signalChan: make(chan os.Signal, 1),
	}
}

// Run -- run ondevice daemon (and return with the exit code of the command)
func (d *Daemon) Run() int {
	d.lock.Path = d.PIDFile
	if err := d.lock.TryLock(); err != nil {
		logg.Fatal("Couldn't acquire lock file")
		return -1
	}
	defer d.lock.Unlock()

	if d.Control != nil {
		d.Control.Start()
		defer d.Control.Stop()
	}

	go d.signalHandler()
	signal.Notify(d.signalChan, syscall.SIGTERM, syscall.SIGINT)

	// TODO implement a sane way to stop this infinite loop (at least SIGTERM, SIGINT or maybe a unix socket call)
	retryDelay := 10 * time.Second
	for !d.shutdown {
		var ws = new(deviceSocket)
		ws.activeTunnels = &d.activeTunnels

		if err := ws.connect(); err != nil {
			retryDelay = d.waitBeforeRetry(retryDelay, err)
		} else {
			d.ws = ws
			ws.Wait()

			if !d.shutdown {
				// connection was successful -> restart after 10sec
				logg.Warning("lost device connection, reconnecting in 10s")
				retryDelay = 10
				time.Sleep(retryDelay * time.Second)
			}
		}
	}

	logg.Info("Stopped ondevice daemon, waiting for remaining tunnels to close (if any...)")
	d.activeTunnels.Wait()

	return 0
}

// Close -- Gracefully stopping this ondevice daemon instance
func (d *Daemon) Close() {
	d.shutdown = true
	if d.Control != nil {
		d.Control.Stop()
	}
	d.ws.Close()
	d.lock.Unlock()
}

// IsOnline -- Returns true if this device is online right now
func (d *Daemon) IsOnline() bool {
	return d.ws != nil && d.ws.IsOnline
}

func (d *Daemon) signalHandler() {
	for true {
		var sig, ok = <-d.signalChan
		if !ok {
			break
		}

		if d.ws == nil {
			// caught the signal before the connection was established -> exit immediately
			logg.Errorf("Caught '%s' signal, exiting", sig)
			os.Exit(1)
		}

		switch sig {
		case syscall.SIGTERM:
			logg.Info("Got SIGTERM, gracefully shutting down...")
		case syscall.SIGINT:
			logg.Info("Got SIGINT, gracefully shutting down...")
		default:
			logg.Warning("Caught unexpected signal: ", sig)
		}

		d.Close()
	}

	logg.Info("Stopping to handle signals")
}

func (d *Daemon) waitBeforeRetry(retryDelay time.Duration, err util.APIError) time.Duration {
	// only abort here if it's an authentication issue
	if err.Code() == util.AuthenticationError {
		logg.Fatal(err)
	}

	// keep retryDelay between 10 and 120sec
	if retryDelay > 120*time.Second {
		retryDelay = 120 * time.Second
	}
	if retryDelay < 10*time.Second {
		retryDelay = 10 * time.Second
	}
	// ... unless we've been rate-limited
	if err.Code() == util.TooManyRequestsError {
		retryDelay = 600 * time.Second
	}

	logg.Debug("device error: ", err)
	logg.Errorf("device error - retrying in %ds", retryDelay/time.Second)

	// sleep to avoid flooding the servers
	time.Sleep(retryDelay)

	// slowly increase retryDelay with each failed attempt
	return time.Duration(float32(retryDelay) * 1.5)
}

func _contains(m *map[string]interface{}, key string) bool {
	_, ok := (*m)[key]
	return ok
}

func _getInt(m *map[string]interface{}, key string) int64 {
	return (*m)[key].(int64)
}

func _getString(m *map[string]interface{}, key string) string {
	if val, ok := (*m)[key]; ok {
		var rc, ok = val.(string)
		if !ok {
			logg.Warningf("Not a string (key %s): %s", key, val)
			return ""
		}
		return rc
	}
	return ""
}
