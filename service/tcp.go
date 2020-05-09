package service

import (
	"net"
	"reflect"

	"github.com/sirupsen/logrus"
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
	logrus.Debug("TCPHandler.Close()")
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
	logrus.Debug("TCPHandler.onEOF()")
	t.Close()
}

func (t *TCPHandler) onData(data []byte) {
	_, err := t.sock.Write(data)
	if err != nil {
		logrus.WithError(err).Error("TCPHandler error: ")
		t.Close()
	}
}

func (t *TCPHandler) receive() {
	buff := make([]byte, 8100)

	for {
		count, err := t.sock.Read(buff)
		if err != nil {
			logrus.WithError(err).Errorf("TCPHandler socket error (%s)", reflect.TypeOf(err))
			break
		}

		if t.tunnel == nil {
			logrus.Fatal("ERROR: TCPHandler.tunnel is null!!!")
			break
		}
		t.tunnel.Write(buff[:count])
	}

	logrus.Debug("TCPHandler: done receiving")
	t.Close()
}

func (t *TCPHandler) self() *ProtocolHandlerBase {
	return &t.ProtocolHandlerBase
}
