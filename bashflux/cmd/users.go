package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var CmdUsers = &cobra.Command{
	Use:   "users",
	Short: "User management",
	Long:  `Manages users in the system (creation, deletition and other system admin)`,
	Run: func(cmdCobra *cobra.Command, args []string) {
		LogUsage("Usage: " + cmdCobra.Short + ". Needs additional commands (see --help)")
	},
}

var CmdCreateUser = &cobra.Command{
	Use:   "create",
	Short: "create <username> <password>",
	Long:  `Creates new user`,
	Run: func(cmdCobra *cobra.Command, args []string) {
		if len(args) == 2 {
			CreateUser(args[0], args[1])
		} else {
			LogUsage(cmdCobra.Short)
		}
	},
}

// Sessions
var CmdSession = &cobra.Command{
	Use:   "tokens",
	Short: "Tokens creation",
	Long:  `Used for tokens manipulation`,
	Run: func(cmdCobra *cobra.Command, args []string) {
		LogUsage("Usage: " + cmdCobra.Short + ". Need additional commands (see --help)")
	},
}

// Init Session
var CmdCreateToken = &cobra.Command{
	Use:   "create",
	Short: "create <username> <password>",
	Long:  `Creates new token`,
	Run: func(cmdCobra *cobra.Command, args []string) {
		if len(args) == 2 {
			CreateToken(args[0], args[1])
		} else {
			LogUsage(cmdCobra.Short)
		}
	},
}

// CreateUser - create user
func CreateUser(user string, pwd string) {
	msg := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, user, pwd)
	url := fmt.Sprintf("%s/users", serverAddr)
	resp, err := httpClient.Post(url, contentType, strings.NewReader(msg))
	FormatResLog(resp, err)
}

// CreateToken - create user token
func CreateToken(user string, pwd string) {
	msg := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, user, pwd)
	url := fmt.Sprintf("%s/tokens", serverAddr)
	resp, err := httpClient.Post(url, contentType, strings.NewReader(msg))
	FormatResLog(resp, err)
}
