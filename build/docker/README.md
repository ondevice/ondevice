ondevice docker image
=====================

This is the official [ondevice.io] CLI docker image.

It's based on [Alpine Linux][alpine-image] and ships with OpenSSH and rsync binaries.

ondevice.io is a service for giving you access to your SSH servers even if they're hidden away behind a NAT or firewall (e.g. your collection of Raspberry PIs or a NAS/server at home) - no need for setting up port forwarding.


## Usage

There are two sides to `ondevice` tunnels: the client and the device it connects to. Have a look at the following sections for ways to set them up.

### on the device

To spin up the `ondevice daemon`, use something like this:

```sh
docker run -d $sshAddr ondevice/ondevice daemon
```

ondevice expects a running SSH server at `ssh:22`, so instead of the above `$sshAddr` use one of:

- `--link=$sshContainer:ssh` if your SSH server runs in another docker container
- `-e SSH_ADDR=$host:$port` to specify the IP/hostname and port of the SSH server
- `--network=host -e SSH_ADDR=127.0.0.1:22` to use the host's network stack (and give the container access to the host's loopback interface)  
  Note that this will only work on Linux hosts, as Docker on macOS/Windows run inside a virtual machine with their own network stack
- `-e SSH_PASSWORD=<password for the 'ondevice' user>`  
  When this variable is present, instead of tunneling incoming connections to a real SSH server, it'll start the builtin `sshd` (and set the password of the `ondevice` user to the value of `$SSH_PASSWORD`).  
  This is mainly intended for testing purposes.


Let's give it a try:

```
$ docker run --rm -d -e SSH_ADDR=192.168.1.10:22 ondevice/ondevice daemon
e7d1c2d61874d8a1792f8495a3329d124a44b9219da874e8ae851acd16056928
```

I've added the `--rm` flag here because we're only testing things out (using the preconfigured `demo` account).  
The container ID is `e7d1c2...`. Let's find out the devId (which we'll need to connect to the device later on)

```
$ docker exec --user=ondevice e7d1c2 ondevice status
Device:
  devID:  demo.ci6lip
  state:  online
  version:  0.5

Client:
  version:  0.5
```

Ok, our new device is called `demo.ci6lip`.  
Each device initially gets a randomly assigned devID (cattle, not pets, right), but feel free to rename them later.


Note: `ondevice status` uses a UNIX socket to communicate with the local ondevice daemon (that's why we run it in the same container).  
Alternatively you could create a new container and use `--volumes-from`

```
$ docker run --rm --volumes-from=e7d1c2 ondevice/ondevice status
Device:
  devID:  demo.ci6lip
  state:  online
  version:  0.5

Client:
  version:  0.5
```


### On the client

Ok, now that we've set up the `ondevice daemon`, let's SSH into it.  
You can do that from any other computer in the world (of course only if they belong to the same ondevice.io user)

On your client PC, run:

```
$ docker run --rm -ti ondevice/ondevice ssh manuel@ci6lip
The authenticity of host 'ondevice:demo.ci6lip (<no hostip for proxy command>)' can't be established.
ECDSA key fingerprint is SHA256:3b0e82559b1b71d18dc66ef1f3ffc0609648f2f8/a8.
Are you sure you want to continue connecting (yes/no)? yes
Warning: Permanently added 'ondevice:demo.ci6lip' (ECDSA) to the list of known hosts.
Password:
Last login: Wed Nov  1 18:29:04 2017 from 192.168.1.10
Manuels-MacBook-Pro:~ manuel$
```

## Environment variables

We've set up the image with credentials for the ondevice.io `demo` user (which has certain restrictions but allows you to give things
a try before creating your own user account).  
Once you've done that, use your own credentials with `ONDEVICE_USER` and `ONDEVICE_AUTH`

- `ONDEVICE_USER` (default: `demo`)  
  Your ondevice.io user name
- `ONDEVICE_AUTH` (default: `5h42l5xylznw`)  
  Your ondevice.io auth key
- `SSH_ADDR` (default: `ssh:22`)  
  SSH server `address:port`
- `SSH_PASSWORD`
  When present, starts the builtin `sshd`, updates the `ondevice` user's password and sets `SSH_ADDR=localhost:22`.  
  This is meant for testing only (that way you won't need to set up SSH_ADDR).
  Use `ondevice ssh ondevice@$devId` on the client to connect to the container's SSH server


[alpine-image]: https://hub.docker.com/_/alpine/
[ondevice.io]: https://ondevice.io/
