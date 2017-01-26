package service

import (
	"log"

	"github.com/ondevice/ondevice/tunnel"
)

// ProtocolHandler -- Service protocol implementations
type ProtocolHandler struct {
	tunnel *tunnel.Tunnel

	Connect func() error
	OnData  func(data []byte)
	OnEOF   func()
	Receive func()
}

// GetProtocolHandler -- Get
func GetProtocolHandler(name string) *ProtocolHandler {
	var rc *ProtocolHandler
	switch name {
	case "echo":
		rc = NewEchoHandler()
	case "ssh":
		rc = NewTCPHandler()
	}

	if rc.Connect != nil {
		err := rc.Connect()
		if err != nil {
			log.Print("GetProtocolHandler error: ", err)
			return nil
		}
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
	p.tunnel = new(tunnel.Tunnel)
	p.tunnel.OnEOF = p.OnEOF
	p.tunnel.OnData = p.OnData

	err := tunnel.Accept(p.tunnel, tunnelID, brokerURL)
	if err == nil {
		if p.Receive != nil {
			go p.Receive()
		}
	}
}
