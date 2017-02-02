package service

// EchoHandler -- ProtocolHandler implementation that simply returns the data it receives
type EchoHandler struct {
	ProtocolHandlerBase
}

// NewEchoHandler -- Create an EchoHandler instance
func NewEchoHandler() ProtocolHandler {
	return new(EchoHandler)
}

func (e *EchoHandler) connect() error {
	return nil /* nop */
}

func (e *EchoHandler) onData(data []byte) {
	e.tunnel.Write(data)
}

func (e *EchoHandler) onEOF() {
	e.tunnel.SendEOF()
}

func (e *EchoHandler) receive() {
	/* nop */
}

func (e *EchoHandler) self() *ProtocolHandlerBase {
	return &e.ProtocolHandlerBase
}
