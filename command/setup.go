package command

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/ondevice/ondevice-cli/config"
	"github.com/ondevice/ondevice-cli/rest"
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

// Run -- implements 'ondevice login'
func (l SetupCmd) Run(args []string) int {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("User: ")
	user, _ := reader.ReadString('\n')
	user = strings.TrimSpace(user)

	fmt.Printf("Auth: ")
	auth, err := gopass.GetPasswd()
	if err != nil {
		log.Fatal(err)
	}

	roles, err := rest.GetKeyInfo(rest.CreateAuth(user, string(auth)))
	if err != nil {
		log.Fatal(err)
	}

	for i := range roles {
		role := roles[i]
		if role == "client" {
			log.Print("updating client auth")
			config.SetValue("client", "user", user)
			config.SetValue("client", "auth", string(auth))
		} else if role == "device" {
			log.Print("updating device auth")
			config.SetValue("device", "user", user)
			config.SetValue("device", "auth", string(auth))
		}
	}

	return 0
}
