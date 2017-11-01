package service

import (
	"os"

	"github.com/ondevice/ondevice/logg"
	"github.com/ondevice/ondevice/tunnel"
)

// ProtocolHandlerBase -- ProtocolHandler base struct
type ProtocolHandlerBase struct {
	tunnel *tunnel.Tunnel
}

// ProtocolHandler -- ProtocolHandler interface
type ProtocolHandler interface {
	self() *ProtocolHandlerBase

	connect() error
	receive()

	onData(data []byte)
	onEOF()
}

// GetProtocolHandler -- Get
func GetProtocolHandler(name string) ProtocolHandler {
	var rc ProtocolHandler
	switch name {
	case "echo":
		rc = NewEchoHandler()
	case "ssh":
		// TODO This is only a temporary (and undocumented) way to get ssh working
		//   in the docker image. As soon as we have proper support for services,
		//   this will be removed (and the docker image will translate the SSH_ADDR
		//   variable to a matching `ondevice service ...` call)
		//   so don't rely on it working (except for the docker image of course)!
		var addr = os.Getenv("SSH_ADDR")
		if addr == "" {
			addr = "127.0.0.1:22"
		}
		rc = NewTCPHandler(addr)
	default:
		logg.Errorf("Unsupported protocol: '%s'", name)
		return nil
	}

	err := rc.connect()
	if err != nil {
		logg.Error("GetProtocolHandler error: ", err)
		return nil
	}

	p := rc.self()
	p.tunnel = new(tunnel.Tunnel)
	p.tunnel.DataListeners = append(p.tunnel.DataListeners, rc.onData)
	p.tunnel.EOFListeners = append(p.tunnel.EOFListeners, rc.onEOF)

	return rc
}

// GetServiceHandler -- Get the ProtocolHandler for a given service
func GetServiceHandler(svc string, protocol string) ProtocolHandler {
	// TODO implement actual services
	if svc != protocol {
		logg.Errorf("protocol/service mismatch: svc=%s, protocol=%s", svc, protocol)
		return nil
	}

	return GetProtocolHandler(protocol)
}

// Start -- Start the tunnel handler
func Start(p ProtocolHandler, tunnelID string, brokerURL string) {
	go run(p, tunnelID, brokerURL)
}

func run(p ProtocolHandler, tunnelID string, brokerURL string) {
	data := p.self()

	err := tunnel.Accept(data.tunnel, tunnelID, brokerURL)
	if err != nil {
		logg.Error("Accepting tunnel failed: ", err)
	} else {
		go p.receive()
	}
}
