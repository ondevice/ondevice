
all: deps
	go build

deps:
	glide install

build-alpine:
	rm -rf target/alpine/

	# build the alpine binary inside docker
	docker build -f build/alpine/Dockerfile -t ondevice-alpine-build .

	# create a temporary container from the image to use for `docker cp`
	$(eval CONTAINER_ID := $(shell docker run --rm -d ondevice-alpine-build sleep 30))
	@[ -n "$(CONTAINER_ID)" ] || (echo "ERROR! Couldn't start temporary container!"; false)
	# copy build result
	mkdir -p target/alpine/
	docker cp "$(CONTAINER_ID):/build/ondevice" ./target/alpine/ondevice

	# clean up
	docker stop "$(CONTAINER_ID)" || true
	#docker rmi ondevice-alpine-build


install:
	mkdir -p $(DESTDIR)/usr/bin/
	install -o root -g root ./ondevice $(DESTDIR)/usr/bin/
