#!/bin/bash -e

CONFFILE=/var/lib/ondevice/ondevice.conf
PIDFILE=/var/run/ondevice/ondevice.pid
SOCKFILE=/var/run/ondevice/ondevice.sock
LOGFILE=/var/log/ondevice/ondevice.log

. /etc/default/ondevice-daemon

export ONDEVICE_USER
export ONDEVICE_AUTH

exec nohup ondevice daemon --conf="$CONFFILE" --pidfile="$PIDFILE" --sock="$SOCKFILE" >> "$LOGFILE" 2>&1
