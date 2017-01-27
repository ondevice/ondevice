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

const _longLoginHelp = `ondevice login

Log in to the ondevice.io service

Example:

$ ondevice login
User: <enter your user name>
Auth: <enter your credentials>
`

// SetupCmd -- `ondevice login` implementation
type SetupCmd struct{}

func (l SetupCmd) args() string {
	return ""
}

func (l SetupCmd) longHelp() string {
	return _longLoginHelp
}

func (l SetupCmd) shortHelp() string {
	return "Log in to the ondevice.io service"
}

func (l SetupCmd) run(args []string) int {
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

	for i := range roles {
		role := roles[i]
		if role == "client" {
			logg.Info("updating client auth")
			config.SetValue("client", "user", user)
			config.SetValue("client", "auth", string(auth))
		} else if role == "device" {
			logg.Info("updating device auth")
			config.SetValue("device", "user", user)
			config.SetValue("device", "auth", string(auth))
		}
	}

	return 0
}
