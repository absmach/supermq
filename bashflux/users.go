package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var cmdUsers = []cobra.Command{
	cobra.Command{
		Use:   "create",
		Short: "create <username> <password>",
		Long:  `Creates new user`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 2 {
				CreateUser(args[0], args[1])
			} else {
				LogUsage(cmd.Short)
			}
		},
	},
	cobra.Command{
		Use:   "token",
		Short: "token <username> <password>",
		Long:  `Creates new token`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 2 {
				CreateToken(args[0], args[1])
			} else {
				LogUsage(cmd.Short)
			}
		},
	},
}

func NewUsersCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "users",
		Short: "User management",
		Long:  `Manages users in the system (creation, deletition and other system admin)`,
		Run: func(cmd *cobra.Command, args []string) {
			LogUsage(cmd.Short)
		},
	}

	for i, _ := range cmdUsers {
		cmd.AddCommand(&cmdUsers[i])
	}

	return &cmd
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
