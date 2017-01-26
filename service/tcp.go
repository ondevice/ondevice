package service

import (
	"net"

	"github.com/ondevice/ondevice/logg"
)

// TCPHandler -- protocol handler connecting to a tcp server
type TCPHandler struct {
	ProtocolHandler

	sock net.Conn
	addr string
}

// NewTCPHandler -- Create new TCPHandler
func NewTCPHandler() *ProtocolHandler {
	rc := new(TCPHandler)
	rc.addr = "127.0.0.1:22"
	rc.OnData = rc.onTunnelData
	rc.OnEOF = rc.onEOF
	rc.Connect = rc.connect
	rc.Receive = rc.receive

	return &rc.ProtocolHandler
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
	// TODO close socket + tunnel
	t.sock.Close()
	t.tunnel.Close()
}

func (t *TCPHandler) onTunnelData(data []byte) {
	_, err := t.sock.Write(data)
	if err != nil {
		logg.Error("TCPHandler error: ", err)
		t.tunnel.Close()
		t.sock.Close()
	}
}

func (t *TCPHandler) receive() {
	buff := make([]byte, 8192)

	for {
		count, err := t.sock.Read(buff)
		if err != nil {
			logg.Error("TCPHandler socket error: ", err)
			t.tunnel.Close()
			t.sock.Close()
			return
		}

		if t.tunnel == nil {
			logg.Fatal("ERROR: TCPHandler.tunnel is null!!!")
		}
		t.tunnel.Write(buff[:count])
	}
}
