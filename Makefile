#
# ondevice Makefile
#
# contains targets for building and packaging the ondevice CLI
#
# travis-ci.org will run `make package` (after

ARCH=$(shell uname -m)
GO_IMAGE=golang:1.9-stretch
VERSION=0.5.1

# Version suffix:
# - empty if TRAVIS_TAG is set
# - +build$n if TRAVIS_BUILD_NUMBER is set
# - "-local" otherwise
VERSION_SUFFIX:=$(if $(TRAVIS_TAG),,$(if $(TRAVIS_BUILD_NUMBER),+build$(TRAVIS_BUILD_NUMBER),-local))

all:
	@mkdir -p target/
	go build -ldflags '-X github.com/ondevice/ondevice/config.version=$(VERSION)' -o target/ondevice ondevice.go

clean:
	rm -rf target/

# prints variables and their values
vars:
	@echo 'Arch: "$(ARCH)"'
	@echo 'Version: "$(VERSION)"'
	@echo 'Suffix: "$(VERSION_SUFFIX)"'

deps:
	glide install

# builds all the release artifacts
package: package-deb package-linux build-docker


#
# ondevice/ondevice docker image
#
build-docker:
	docker build -f build/docker/Dockerfile -t ondevice/ondevice .


#
# .deb files
#
package-deb: package-deb-amd64 package-deb-i386 package-deb-armhf

package-deb-amd64:
	$(MAKE) _package-deb ARCH=amd64 GOARCH=amd64 SOURCE_IMAGE=amd64/$(GO_IMAGE)

package-deb-i386:
	$(MAKE) _package-deb ARCH=i386 GOARCH=386 SOURCE_IMAGE=i386/$(GO_IMAGE)

package-deb-armhf:
	$(MAKE) _package-deb ARCH=armhf GOARCH=arm SOURCE_IMAGE=$(GO_IMAGE) BUILD_ARGS=--host-arch=armhf


_package-deb:
	@echo "\n============\nPackaging: debian $(ARCH)\n============\n" >&2
	# builds and packages the i386,amd64 and armhf ondevice debian packages (as well as ondevice-daemon)
	docker build -f build/deb/Dockerfile '--build-arg=SOURCE_IMAGE=$(SOURCE_IMAGE)' '--build-arg=GOARCH=$(GOARCH)' '--build-arg=BUILD_ARGS=$(BUILD_ARGS)' '--build-arg=VERSION=$(VERSION)' '--build-arg=VERSION_SUFFIX=$(VERSION_SUFFIX)' -t ondevice/package-deb-$(ARCH) .

	# extract artefacts
	rm -rf 'target/deb/$(ARCH)'; mkdir -p target/deb/
	CONTAINER="$$(docker run --rm -d ondevice/package-deb-$(ARCH) sleep 60)"; docker cp "$$CONTAINER:/target" 'target/deb/$(ARCH)'

#
# Linux .tgz
#
# TODO this build runs as root (and might create the target as root breaking package-deb running later)
package-linux: package-linux-amd64 package-linux-i386 package-linux-armhf
package-linux-amd64:
	$(MAKE) _package-linux GOARCH=amd64
package-linux-armhf:
	$(MAKE) _package-linux GOARCH=arm
package-linux-i386:
	$(MAKE) _package-linux GOARCH=386

_package-linux:
	@echo "\n============\nPackaging: linux $(GOARCH)\n============\n" >&2
	docker run --rm -ti -v "$(PWD):/go/src/github.com/ondevice/ondevice/" "$(GO_IMAGE)" env GOARCH="$(GOARCH)" VERSION="$(VERSION)" VERSION_SUFFIX="$(VERSION_SUFFIX)" /go/src/github.com/ondevice/ondevice/build/linux/build.sh
