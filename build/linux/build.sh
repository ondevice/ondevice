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

# debian-style architecture, derived from $GOARCH (used in the resulting .tgz filename)
ARCH="$(echo "$GOARCH"| sed 's/^386$/i386/;s/^arm$/armhf/')"

# install glide
curl https://glide.sh/get | sh


# build ondevice + prepare target dir
cd "$BASEDIR"
glide update
mkdir -p /tmp/build/usr/lib/systemd/system/ /tmp/build/usr/bin/ "$BASEDIR/target/"
go build -ldflags "-X github.com/ondevice/ondevice/config.version=$VERSION" -o /tmp/build/usr/bin/ondevice ondevice.go
install build/linux/ondevice-daemon.service /tmp/build/usr/lib/systemd/system/

# create .tgz
cd /tmp/build
tar cfz $BASEDIR/target/ondevice-linux_${VERSION}_${ARCH}.tgz ./usr
