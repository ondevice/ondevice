package command

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/ondevice/ondevice/api"
	"github.com/ondevice/ondevice/config"
	"github.com/ondevice/ondevice/logg"
)

func loginRun(args []string) int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("User: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	fmt.Printf("Auth: ")
	auth, err := gopass.GetPasswd()
	if err != nil {
		logg.Fatal(err)
	}

	roles, err := api.GetKeyInfo(api.CreateAuth(user, string(auth)))
	if err != nil {
		logg.Fatal(err)
	}

	for _, role := range roles {
		if role == "client" {
			logg.Info("updating client auth")
			if err := config.SetAuth("client", user, string(auth)); err != nil {
				logg.Fatal("Failed to set client auth: ", err)
			}
		} else if role == "device" {
			logg.Info("updating device auth")
			if err := config.SetAuth("device", user, string(auth)); err != nil {
				logg.Fatal("Failed to set device auth: ", err)
			}
		}
	}

	return 0
}

// LoginCommand -- implements `ondevice login`
var LoginCommand = BaseCommand{
	Arguments: "",
	ShortHelp: "Log in to the ondevice.io service",
	RunFn:     nil,
	LongHelp: `$ ondevice login

Log in to the ondevice.io service using one of your API keys.

Example:
  $ ondevice login
  User: <enter your user name>
  Auth: <enter your credentials>
`,
}
