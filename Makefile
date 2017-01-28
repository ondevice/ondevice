
all: deps
	go build

deps:
	go get github.com/gorilla/websocket
	go get github.com/howeyc/gopass
	go get github.com/jessevdk/go-flags
	go get gopkg.in/ini.v1

install:
	mkdir -p $(DESTDIR)/usr/bin/
	install -o root -g root ./ondevice $(DESTDIR)/usr/bin/
