#!/bin/bash -e

if [ -z "$PROJECT_DIR" ]; then
	BASEDIR=/go/src/github.com/ondevice/ondevice/
fi

if ! [ -d "$BASEDIR" ]; then
	echo "ERROR: Couldn't find '$BASEDIR'" >&2
	false
fi
if [ -z "$GOARCH" ]; then
	echo 'ERROR: Missing $GOARCH' >&2
	false
fi
if [ -z "$VERSION" ]; then
	echo 'ERROR: Missing $VERSION' >&2
	false
fi
if [ -d /tmp/build/ ]; then
	echo "ERROR: Target dir already exists!" >&2
	false
fi

# append $VERSION_SUFFIX if set
if [ -n "$VERSION_SUFFIX" ]; then
	VERSION="${VERSION}${VERSION_SUFFIX}"
fi

# debian-style architecture, derived from $GOARCH (used in the resulting .tgz filename)
ARCH="$(echo "$GOARCH"| sed 's/^386$/i386/;s/^arm$/armhf/')"

# install glide
curl https://glide.sh/get | sh


# build ondevice + prepare target dir
cd "$BASEDIR"
glide update
mkdir -p /tmp/build/lib/systemd/system/ /tmp/build/usr/lib/tmpfiles.d/ /tmp/build/usr/bin/ /tmp/build/etc/init.d/ "$BASEDIR/target/"
go build -ldflags "-X github.com/ondevice/ondevice/config.version=$VERSION" -o /tmp/build/usr/bin/ondevice ondevice.go
install -m 0644 build/linux/ondevice-daemon.service /tmp/build/lib/systemd/system/
install -m 0644 build/linux/ondevice-daemon.tmpfile /tmp/build/usr/lib/tmpfiles.d/ondevice-daemon.conf
install -m 0755 build/linux/ondevice-daemon.init.d /tmp/build/etc/init.d/ondevice-daemon

# create .tgz
cd /tmp/build
tar cfz $BASEDIR/target/ondevice_${VERSION}_linux-${ARCH}.tgz ./
