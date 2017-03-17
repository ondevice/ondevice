package command

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
	"github.com/ondevice/ondevice/tunnel"
	"github.com/ondevice/ondevice/util"
)

// PipeCmd -- `ondevice :pipe` implementation
type PipeCmd struct {
	tunnel *tunnel.Tunnel

	sentEOF bool

	reader *bufio.Reader
	writer *bufio.Writer
}

func (p *PipeCmd) args() string {
	return "<devId> <service>"
}

func (p *PipeCmd) longHelp() string {
	logg.Fatal("implement me")
	return ""
}

func (p *PipeCmd) shortHelp() string {
	return "Pipes a device's service to stdout/stdin"
}

func (p *PipeCmd) run(args []string) int {
	// parse arguments
	if len(args) < 1 {
		logg.Fatal("Missing devId")
	}
	if len(args) < 2 {
		logg.Fatal("Missing service name")
	}

	devID := args[0]
	service := args[1]

	auth, err := api.CreateClientAuth()
	if err != nil {
		logg.Fatal("Missing client credentials")
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
	p.tunnel = &c
	c.CloseListeners = append(c.CloseListeners, p.onClose)
	c.DataListeners = append(c.DataListeners, p.onData)
	c.ErrorListeners = append(c.ErrorListeners, p.onError)
	if e := tunnel.Connect(&c, devID, service, service, auth); e != nil {
		logg.FailWithAPIError(e)
	}

	buff := make([]byte, 8100)
	var count int

	for {
		if count, err = p.reader.Read(buff); err != nil {
			if count == 0 && err == io.EOF {
				// we can't simply call c.Close() here because the other side might still
				// send data. A simple test would be (assuming the device has the 'echo' service enabled):
				//   echo hello | ondevice :pipe <dev> echo
				p.sentEOF = true
				c.SendEOF()
				break
			} else {
				logg.Fatal("error reading from stdin: ", err)
			}
		}

		c.Write(buff[:count])
	}

	c.Wait()
	return 0
}

func (p *PipeCmd) onClose() {
	// odds are run() is currently blocking in the p.reader.Read(). Close stdin to
	// allow it to return gracefully
	logg.Debug("Pipe: Connection closed")
	os.Stdin.Close()
}

func (p *PipeCmd) onError(err util.APIError) {
	if err.Code() != util.OtherError {
		logg.FailWithAPIError(err)
	} else if !p.sentEOF {
		logg.Fatal("Lost connection: ", err)
	}
	p.tunnel.Close()
}

// OnMessage -- Handles incoming WebSocket messages
func (p *PipeCmd) onData(data []byte) {
	p.writer.Write(data)
	p.writer.Flush()
}
