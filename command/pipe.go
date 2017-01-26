package command

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/tunnel"
)

// PipeCmd -- `ondevice :pipe` implementation
type PipeCmd struct {
	ws *websocket.Conn

	sentEOF bool

	reader *bufio.Reader
	writer *bufio.Writer
}

func (p *PipeCmd) args() string {
	return "<devId> <service>"
}

func (p *PipeCmd) longHelp() string {
	log.Fatal("implement me")
	return ""
}

func (p *PipeCmd) shortHelp() string {
	return "Pipes a device's service to stdout/stdin"
}

// Run -- implements `ondevice pipe`
func (p *PipeCmd) Run(args []string) int {
	// parse arguments
	if len(args) < 1 {
		log.Fatal("Missing devId")
	}
	if len(args) < 2 {
		log.Fatal("Missing service name")
	}

	devID := args[0]
	service := args[1]

	auth, err := api.CreateClientAuth()
	if err != nil {
		log.Fatal("Missing client credentials")
	}

	if strings.Contains(devID, ".") {
		parts := strings.SplitN(devID, ".", 2)
		var user, pwd string
		if user, pwd, err = config.GetClientUserAuth(parts[0]); err == nil {
			devID = parts[1]
			auth = api.CreateAuth(user, pwd)
		}
	}

	p.writer = bufio.NewWriter(os.Stdout)
	p.reader = bufio.NewReader(os.Stdin)

	// initiate connection
	c := tunnel.Tunnel{}
	c.OnError = p.onError
	c.OnData = p.onData
	if err = tunnel.Connect(&c, devID, service, service, auth); err != nil {
		log.Fatal(err)
	}

	buff := make([]byte, 8192)
	var count int

	for {
		if count, err = p.reader.Read(buff); err != nil {
			if count == 0 && err == io.EOF {
				// we can't simply call c.Close() here because the other side might still
				// send data. A simple test would be (assuming the device has the 'echo' service enabled):
				//   echo hello | ondevice pipe <dev> echo
				p.sentEOF = true
				c.CloseWrite()
				break
			} else {
				log.Fatal("error reading from stdin: ", err)
			}
		}

		c.Write(buff[:count])
	}

	c.Wait()
	return 0
}

func (p *PipeCmd) onError(err error) {
	if !p.sentEOF {
		log.Fatal("Lost connection")
	}
	p.ws.Close()
}

// OnMessage -- Handles incoming WebSocket messages
func (p *PipeCmd) onData(data []byte) {
	p.writer.Write(data)
	p.writer.Flush()
}
