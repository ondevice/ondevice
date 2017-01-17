package command

import (
	"bufio"
	"log"
	"os"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice-cli/tunnel"
)

// PipeCmd -- `ondevice pipe` implementation
type PipeCmd struct {
	ws *websocket.Conn

	reader *bufio.Reader
	writer *bufio.Writer
}

func (p PipeCmd) args() string {
	return "<devId> <service>"
}

func (p PipeCmd) longHelp() string {
	log.Fatal("implement me")
	return ""
}

func (p PipeCmd) shortHelp() string {
	return "Pipes a device's service to stdout/stdin"
}

// Run -- implements `ondevice pipe`
func (p PipeCmd) Run(args []string) {
	// parse arguments
	if len(args) < 1 {
		log.Fatal("Missing devId")
	}
	if len(args) < 2 {
		log.Fatal("Missing service name")
	}

	devID := args[0]
	service := args[1]

	// initiate connection
	c, err := tunnel.Connect(devID, service, service)
	c.OnData = p.onData
	if err != nil {
		log.Fatal(err)
	}

	p.writer = bufio.NewWriter(os.Stdout)
	p.reader = bufio.NewReader(os.Stdin)
	buff := make([]byte, 8192)
	var count int

	for {
		if count, err = p.reader.Read(buff); err != nil {
			log.Fatal("error reading from stdin: ", err)
		}

		c.Write(buff[:count])
	}
}

// OnMessage -- Handles incoming WebSocket messages
func (p *PipeCmd) onData(data []byte) {
	p.writer.Write(data)
	p.writer.Flush()
}
