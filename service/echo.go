package service

// EchoHandler -- ProtocolHandler implementation that simply returns the data it receives
type EchoHandler struct {
	ProtocolHandler
}

func (e *EchoHandler) onRemoteEOF() {
	e.t.Close()
}

// OnData -- Incoming message handler - simply sending them back to their origin
func (e *EchoHandler) onData(data []byte) {
	e.t.Write(data)
}
