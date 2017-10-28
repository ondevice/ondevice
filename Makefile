
all: deps
	go build

deps:
	glide install

build-alpine:
	# create the (temporary) build image
	docker build -f build/alpine/Dockerfile -t ondevice-alpine-build .

	# create a temporary container from the image to use for `docker cp`
	$(eval CONTAINER_ID := $(shell docker run --rm -d ondevice-alpine-build sleep 30))
	@[ -n "$(CONTAINER_ID)" ] || echo "ERROR! Couldn't start temporary container!"
	# copy build result
	mkdir -p target/alpine/
	docker cp "$(CONTAINER_ID):/build/ondevice" ./target/alpine/ondevice

	# clean up
	docker stop "$(CONTAINER_ID)"
	docker rmi ondevice-docker-build


install:
	mkdir -p $(DESTDIR)/usr/bin/
	install -o root -g root ./ondevice $(DESTDIR)/usr/bin/
