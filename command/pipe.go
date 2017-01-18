package command

import (
	"bufio"
	"io"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/ondevice/ondevice-cli/config"
	"github.com/ondevice/ondevice-cli/rest"
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
func (p PipeCmd) Run(args []string) int {
	// parse arguments
	if len(args) < 1 {
		log.Fatal("Missing devId")
	}
	if len(args) < 2 {
		log.Fatal("Missing service name")
	}

	// TODO if devID is qualified and it's another user, use other client credentials if possible
	devID := args[0]
	service := args[1]

	auth, err := rest.CreateClientAuth()
	if err != nil {
		log.Fatal("Missing client credentials")
	}

	if strings.Contains(devID, ".") {
		parts := strings.SplitN(devID, ".", 2)
		var user, pwd string
		if user, pwd, err = config.GetClientUserAuth(parts[0]); err == nil {
			devID = parts[1]
			auth = rest.CreateAuth(user, pwd)
		}
	}

	// initiate connection
	c, err := tunnel.Connect(devID, service, service, auth)
	if err != nil {
		log.Fatal(err)
	}
	c.OnData = p.onData

	p.writer = bufio.NewWriter(os.Stdout)
	p.reader = bufio.NewReader(os.Stdin)
	buff := make([]byte, 8192)
	var count int

	for {
		if count, err = p.reader.Read(buff); err != nil {
			if count == 0 && err == io.EOF {
				// we can't simply call c.Close() here because the other side might still
				// send data. A simple test would be (assuming the device has the 'echo' service enabled):
				//   echo hello | ondevice pipe <dev> echo
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

// OnMessage -- Handles incoming WebSocket messages
func (p *PipeCmd) onData(data []byte) {
	p.writer.Write(data)
	p.writer.Flush()
}
