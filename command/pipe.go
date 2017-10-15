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

// pipeCommand -- `ondevice pipe` implementation
type pipeCommand struct {
	BaseCommand
	tunnel *tunnel.Tunnel

	sentEOF bool

	reader *bufio.Reader
	writer *bufio.Writer
}

func (p pipeCommand) run(args []string) int {
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
				//   echo hello | ondevice pipe <dev> echo
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

func (p *pipeCommand) onClose() {
	// odds are run() is currently blocking in the p.reader.Read(). Close stdin to
	// allow it to return gracefully
	logg.Debug("Tunnel closed")
	os.Stdin.Close()
}

func (p *pipeCommand) onError(err util.APIError) {
	if err.Code() != util.OtherError {
		logg.FailWithAPIError(err)
	} else if !p.sentEOF {
		logg.Fatal("Lost connection: ", err)
	}
	p.tunnel.Close()
}

// OnMessage -- Handles incoming WebSocket messages
func (p *pipeCommand) onData(data []byte) {
	p.writer.Write(data)
	p.writer.Flush()
}

// PipeCommand -- Implements `ondevice pipe`
var PipeCommand = pipeCommand{
	BaseCommand: BaseCommand{
		Arguments: "<devId> <service>",
		ShortHelp: "Pipes a device's service to stdout/stdin",
		RunFn:     nil, // we're implemnting our own run() method
		LongHelp: `$ ondevice pipe <devId> <service>

Sends data from stdin to the specified service - and prints whatever it gets in
return to stdout.

This command is used internally by 'ondevice ssh' to serve as ssh's ProxyCommand.

Example:
  $ echo hello world | ondevice pipe <devId> echo
  hello world

	Sends 'hello world' to q5dkpm's 'echo' service. The echo service simply returns
	the data it gets back to the sender. Therefore the above command is equivalent
	to simply calling 'echo hello world' (as long as your device is online).
`,
	},
}
