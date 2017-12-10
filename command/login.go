package command

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	flags "github.com/jessevdk/go-flags"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

type loginOpts struct {
	BatchUser string `long:"batch" description:"If set, use that user to login and read the auth key from stdin"`
}

func loginRun(args []string) int {
	var user, auth string
	var opts loginOpts
	var err error
	if _, err = flags.ParseArgs(&opts, args); err != nil {
		logg.Fatal(err)
	}

	reader := bufio.NewReader(os.Stdin)

	if opts.BatchUser != "" {
		user = opts.BatchUser
		// run in batch mode
		if auth, err = reader.ReadString('\n'); err != nil {
			logg.Fatal("Failed to read auth key from stdin: ", err)
		}
		auth = strings.TrimSpace(auth)
	} else {
		fmt.Print("User: ")
		user, err = reader.ReadString('\n')
		if err != nil {
			logg.Fatal("Failed to read user name: ", err)
		}
		user = strings.TrimSpace(user)

		fmt.Printf("Auth: ")
		var authBytes []byte
		if authBytes, err = gopass.GetPasswd(); err != nil {
			logg.Fatal(err)
		}
		auth = string(authBytes)
	}

	info, err := api.GetKeyInfo(api.CreateAuth(user, auth))
	if err != nil {
		logg.Fatal(err)
	}

	if info.Key != "" {
		// the API server wants us to use a different auth key
		// (most likely because the user has used their account password)
		auth = info.Key
	}

	// display any messages the server might have for us
	for _, msg := range info.Messages {
		var parts = strings.SplitN(msg, ":", 2)
		switch parts[0] {
		case "info":
			logg.Info(parts[1])
		case "warn":
			logg.Warning(parts[1])
		case "err":
			logg.Error(parts[1])
		default:
			logg.Info("Got message: ", msg)
		}
	}

	// update auth
	if info.IsType("client") {
		logg.Info("updating client auth")
		if err := config.SetAuth("client", user, string(auth)); err != nil {
			logg.Fatal("Failed to set client auth: ", err)
		}
	}
	if info.IsType("device") {
		logg.Info("updating device auth")
		if err := config.SetAuth("device", user, string(auth)); err != nil {
			logg.Fatal("Failed to set device auth: ", err)
		}
	}

	return 0
}

// LoginCommand -- implements `ondevice login`
var LoginCommand = BaseCommand{
	Arguments: "[--batch=username]",
	ShortHelp: "Log in to the ondevice.io service",
	RunFn:     loginRun,
	LongHelp: `$ ondevice login

Log in to the ondevice.io service using one of your API keys.

Options:
--batch=username
  Run in batch mode, using the given username and reading the authentication key
  from stdin, e.g.:
    echo '5h42l5xylznw'|ondevice login --batch=demo


Example:
  $ ondevice login
  User: <enter your user name>
  Auth: <enter your credentials>
`,
}
