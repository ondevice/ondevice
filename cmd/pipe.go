/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"bufio"
	"io"
	"os"

	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/tunnel"
	"github.com/ondevice/ondevice/util"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// pipeCmd represents the pipe command
type pipeCmd struct {
	cobra.Command

	tunnel *tunnel.Tunnel
	reader *bufio.Reader
	writer *bufio.Writer

	sentEOF bool
}

func init() {
	var c pipeCmd
	c.Command = cobra.Command{
		Use:   "pipe <devId> <service>",
		Short: "pipes a device's service to stdout/stdin",
		Long: `sends data from stdin to the specified service - and prints whatever it gets
in return to stdout.

This command is used internally by 'ondevice ssh' to serve as ssh's ProxyCommand.`,
		Example: `  $ echo hello world | ondevice pipe <devId> echo
  hello world

  Sends 'hello world' to q5dkpm's 'echo' service. The echo service simply returns
  data back to the sender unaltered. Therefore the above command is equivalent
  to simply calling 'echo hello world' (as long as your device is online).`,
		Run:    c.run,
		Hidden: true,
	}
	rootCmd.AddCommand(&c.Command)
}

func (c *pipeCmd) run(cmd *cobra.Command, args []string) {
	// parse arguments
	if len(args) < 1 {
		logrus.Fatal("missing devId")
		return
	}
	if len(args) < 2 {
		logrus.Fatal("missing service name")
		return
	}

	devID := args[0]
	service := args[1]

	auth, err := config.LoadAuth().GetClientAuthForDevice(devID)
	if err != nil {
		logrus.WithError(err).Fatal("missing client credentials")
		return
	}

	c.writer = bufio.NewWriter(os.Stdout)
	c.reader = bufio.NewReader(os.Stdin)

	// initiate connection
	t := tunnel.Tunnel{}
	c.tunnel = &t

	t.CloseListeners = append(t.CloseListeners, c.onClose)
	t.DataListeners = append(t.DataListeners, c.onData)
	t.ErrorListeners = append(t.ErrorListeners, c.onError)
	if e := tunnel.Connect(&t, devID, service, service, auth); e != nil {
		util.FailWithAPIError(e)
	}

	buff := make([]byte, 8100)
	var count int

	for {
		if count, err = c.reader.Read(buff); err != nil {
			if count == 0 && err == io.EOF {
				// we can't simply call c.Close() here because the other side might still
				// send data. A simple test would be (assuming the device has the 'echo' service enabled):
				//   echo hello | ondevice pipe <dev> echo
				c.sentEOF = true
				t.SendEOF()
				break
			} else {
				logrus.WithError(err).Fatal("error reading from stdin")
				return
			}
		}

		t.Write(buff[:count])
	}

	t.Wait()
}

func (*pipeCmd) onClose() {
	// odds are run() is currently blocking in the p.reader.Read(). Close stdin to
	// allow it to return gracefully
	logrus.Debug("tunnel closed")
	os.Stdin.Close()
}

// OnMessage -- Handles incoming WebSocket messages
func (c *pipeCmd) onData(data []byte) {
	c.writer.Write(data)
	c.writer.Flush()
}

func (c *pipeCmd) onError(err util.APIError) {
	if err.Code() != util.OtherError {
		util.FailWithAPIError(err)
	} else if !c.sentEOF {
		logrus.WithError(err).Fatal("lost connection")
	}
	c.tunnel.Close()
}
