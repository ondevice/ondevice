
all: deps
	go build

deps:
	glide install

install:
	mkdir -p $(DESTDIR)/usr/bin/
	install -o root -g root ./ondevice $(DESTDIR)/usr/bin/
