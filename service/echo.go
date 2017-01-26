package service

// EchoHandler -- ProtocolHandler implementation that simply returns the data it receives
type EchoHandler struct {
	ProtocolHandler
}

// NewEchoHandler -- Create an EchoHandler instance
func NewEchoHandler() *ProtocolHandler {
	rc := EchoHandler{}
	rc.OnData = rc.onData
	rc.OnEOF = rc.onRemoteEOF
	return &rc.ProtocolHandler
}

func (e *EchoHandler) onRemoteEOF() {

	e.tunnel.Close()
}

// OnData -- Incoming message handler - simply sending them back to their origin
func (e *EchoHandler) onData(data []byte) {
	e.tunnel.Write(data)
}
