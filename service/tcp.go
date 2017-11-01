package service

import (
	"net"
	"reflect"

	"github.com/ondevice/ondevice/logg"
)

// TCPHandler -- protocol handler connecting to a tcp server
type TCPHandler struct {
	ProtocolHandlerBase

	sock     net.Conn
	addr     string
	isClosed bool
}

// NewTCPHandler -- Create new TCPHandler
func NewTCPHandler(addr string) ProtocolHandler {
	rc := new(TCPHandler)
	rc.addr = addr

	return rc
}

// Close -- Close both connections
func (t *TCPHandler) Close() {
	if t.isClosed {
		return
	}
	logg.Debug("TCPHandler.Close()")
	t.sock.Close()
	t.tunnel.Close()
	t.isClosed = true
}

func (t *TCPHandler) connect() error {
	var err error
	t.sock, err = net.Dial("tcp", t.addr)
	if err != nil {
		return err
	}

	return nil
}

func (t *TCPHandler) onEOF() {
	logg.Debug("TCPHandler.onEOF()")
	t.Close()
}

func (t *TCPHandler) onData(data []byte) {
	_, err := t.sock.Write(data)
	if err != nil {
		logg.Error("TCPHandler error: ", err)
		t.Close()
	}
}

func (t *TCPHandler) receive() {
	buff := make([]byte, 8100)

	for {
		count, err := t.sock.Read(buff)
		if err != nil {
			logg.Errorf("TCPHandler socket error (%s): %s", reflect.TypeOf(err), err)
			break
		}

		if t.tunnel == nil {
			logg.Fatal("ERROR: TCPHandler.tunnel is null!!!")
		}
		t.tunnel.Write(buff[:count])
	}

	logg.Debug("TCPHandler: done receiving")
	t.Close()
}

func (t *TCPHandler) self() *ProtocolHandlerBase {
	return &t.ProtocolHandlerBase
}
