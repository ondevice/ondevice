
all: deps
	go build

deps:
	glide install


build-docker:
	docker build -f build/docker/Dockerfile -t ondevice/ondevice .


install:
	mkdir -p $(DESTDIR)/usr/bin/
	install -o root -g root ./ondevice $(DESTDIR)/usr/bin/
