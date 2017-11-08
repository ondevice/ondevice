
ARCH=$(shell uname -m)

all:
	mkdir -p target/
	go build -o target/ondevice ondevice.go

clean:
	rm -rf target/

deps:
	glide install

build-docker:
	docker build -f build/docker/Dockerfile -t ondevice/ondevice .

package-deb: package-deb-$(ARCH)

package-deb-x86_64: package-deb-amd64
package-deb-amd64:
	$(MAKE) _package-deb ARCH=amd64 GOARCH=amd64 SOURCE_IMAGE=amd64/golang:1.9-stretch

package-deb-i386:
	$(MAKE) _package-deb ARCH=i386 GOARCH=386 SOURCE_IMAGE=i386/golang:1.9-stretch

package-deb-armv7l: package-deb-armhf
package-deb-armhf:
	$(MAKE) _package-deb ARCH=armhf GOARCH=armv6l SOURCE_IMAGE=golang:1.9-stretch


_package-deb:
	# builds and packages the i386+amd64 ondevice debian packages (as well as ondevice-daemon)
	docker build -f build/deb/Dockerfile '--build-arg=ARCH=$(ARCH)' '--build-arg=SOURCE_IMAGE=$(SOURCE_IMAGE)' '--build-arg=GOARCH=$(GOARCH)' -t ondevice/package-deb-$(ARCH) .

	# extract artefacts
	#CONTAINER="$$(docker run --rm -d ondevice/package-deb-$(ARCH) sleep 60)" docker cp "$$CONTAINER:/build.tgz" /tmp/
