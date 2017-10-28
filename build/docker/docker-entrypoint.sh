#!/bin/sh -e

# fix volume permissions
chown -R ondevice:ondevice /home/ondevice/.config/

# anything starting with a / will be run as-is
if echo "$1" | grep -q ^/; then
	exec "$@"
else
	exec su-exec ondevice ondevice "$@"
fi
