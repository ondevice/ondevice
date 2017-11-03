#!/bin/bash -e

CONFFILE=/var/lib/ondevice/ondevice.conf
PIDFILE=/var/run/ondevice/ondevice.pid
SOCKFILE=/var/run/ondevice/ondevice.sock

. /etc/default/ondevice-daemon

export ONDEVICE_USER
export ONDEVICE_AUTH

exec ondevice daemon --conf="$CONFFILE" --pidfile="$PIDFILE" --sock="$SOCKFILE" &>> /var/log/ondevice/ondevice.log
