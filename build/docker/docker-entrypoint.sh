#!/bin/sh -e

# fix volume permissions
chown -R ondevice:ondevice /home/ondevice/.config/

# TODO as soon as `ondevice service` support exists, parse $SSH_ADDR
# and call something like `ondevice service set ssh 'addr=$SSH_ADDR'

# anything starting with a / will be run as-is
if echo "$1" | grep -q ^/; then
	exec "$@"
else
	if [ -n "$SSH_PASSWORD" ]; then
		# User wants to use the builtin sshd
		export SSH_ADDR=localhost:22
		echo "ondevice:$SSH_PASSWORD" | chpasswd
		ssh-keygen -A

		#start sshd
		/usr/sbin/sshd -e
	fi

	exec su-exec ondevice ondevice "$@"
fi
