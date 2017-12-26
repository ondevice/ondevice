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

cd "$BASEDIR"

mkdir -p /tmp/build/usr/lib/systemd/system/ /tmp/build/usr/bin/ "$BASEDIR/target/"
go build -ldflags "-X github.com/ondevice/ondevice/config.version=$VERSION" -o /tmp/build/usr/bin/ondevice ondevice.go
install build/linux/ondevice.service /tmp/build/usr/lib/systemd/system/

cd /tmp/build
tar cfz $BASEDIR/target/ondevice_linux_${VERSION}_${GOARCH}.tgz ./usr
