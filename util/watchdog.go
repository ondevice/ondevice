package util

import "time"

// Watchdog -- Periodically call a function until stop() has been called (or the callback returned an error)
type Watchdog struct {
	onTimeout func()
	timer     *time.Timer
	timeout   time.Duration
}

// NewWatchdog -- Create and start a new Watchdog
func NewWatchdog(timeout time.Duration, callback func()) *Watchdog {
	return &Watchdog{
		timeout:   timeout,
		timer:     time.AfterFunc(timeout, callback),
		onTimeout: callback,
	}
}

// Kick -- Trigger the Watchdog (reset the timeout)
func (w *Watchdog) Kick() {
	w.timer.Stop()
	w.timer.Reset(w.timeout)
}

// Stop -- Stop the Watchdog
func (w *Watchdog) Stop() {
	w.timer.Stop()
}
