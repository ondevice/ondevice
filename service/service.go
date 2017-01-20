package service

import (
	"log"

	"github.com/ondevice/ondevice/tunnel"
)

// ProtocolHandler -- Service protocol implementations
type ProtocolHandler struct {
	t *tunnel.Tunnel

	OnData   func(data []byte)
	OnEOF    func()
	RunLocal func()
}

// GetProtocolHandler -- Get
func GetProtocolHandler(name string) *ProtocolHandler {
	var rc *ProtocolHandler
	switch name {
	case "echo":
		e := new(EchoHandler)
		e.OnData = e.onData
		e.OnEOF = e.onRemoteEOF
		rc = &e.ProtocolHandler
	}

	return rc
}

// GetServiceHandler -- Get the ProtocolHandler for a given service
func GetServiceHandler(svc string, protocol string) *ProtocolHandler {
	// TODO implement actual services
	if svc != protocol {
		log.Printf("protocol/service mismatch: svc=%s, protocol=%s", svc, protocol)
		return nil
	}

	return GetProtocolHandler(protocol)
}

// Start -- Start the tunnel handler
func (p *ProtocolHandler) Start(tunnelID string, brokerURL string) {
	go p.run(tunnelID, brokerURL)
}

func (p *ProtocolHandler) run(tunnelID string, brokerURL string) {
	p.t = new(tunnel.Tunnel)
	p.t.OnEOF = p.OnEOF
	p.t.OnData = p.OnData
	tunnel.Accept(p.t, tunnelID, brokerURL)
}
