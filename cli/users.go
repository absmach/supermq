// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"

	mfxsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/spf13/cobra"
)

var cmdUsers = []cobra.Command{
	{
		Use:   "create <username> <password> <user_auth_token>",
		Short: "Create user",
		Long:  `Creates new user`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 || len(args) > 3 {
				logUsage(cmd.Use)
				return
			}
			if len(args) == 2 {
				args = append(args, "")
			}

			user := mfxsdk.User{
				Email:    args[0],
				Password: args[1],
			}
			id, err := sdk.CreateUser(args[2], user)
			if err != nil {
				logError(err)
				return
			}

			logCreated(id)
		},
	},
	{
		Use:   "get [all | email <email_address> | metadata <metadata_json_string> | <user_id> ] <user_auth_token>",
		Short: "Get users",
		Long: `Get all users, group by id, group by email or group by metadata.
		all - lists all users
		email <email_address> - list all users with <email_address> 
		metadata <metadata_json_string> - list all users with <metadata_json_string>
		<user_id> - shows user with provided <user_id>`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				logUsage(cmd.Use)
				return
			}
			pageMetadata := mfxsdk.PageMetadata{
				Email:    "",
				Offset:   uint64(Offset),
				Limit:    uint64(Limit),
				Metadata: make(map[string]interface{}),
			}
			if args[0] == "all" {
				l, err := sdk.Users(args[1], pageMetadata)
				if err != nil {
					logError(err)
					return
				}
				logJSON(l)
				return
			}
			if args[0] == "email" {
				pageMetadata.Email = args[1]
				l, err := sdk.Users(args[2], pageMetadata)
				if err != nil {
					logError(err)
					return
				}
				logJSON(l)
				return
			}
			if args[0] == "metadata" {
				var metadata map[string]interface{}
				if err := json.Unmarshal([]byte(args[1]), &metadata); err != nil {
					logError(err)
					return
				}
				pageMetadata.Metadata = metadata
				l, err := sdk.Users(args[2], pageMetadata)
				if err != nil {
					logError(err)
					return
				}
				logJSON(l)
				return
			}
			if len(args) > 2 {
				logUsage(cmd.Use)
				return
			}
			u, err := sdk.User(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(u)
		},
	},
	{
		Use:   "token <username> <password>",
		Short: "Get token",
		Long:  `Generate new token`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			user := mfxsdk.User{
				Email:    args[0],
				Password: args[1],
			}
			token, err := sdk.CreateToken(user)
			if err != nil {
				logError(err)
				return
			}

			logCreated(token)

		},
	},
	{
		Use:   "update <JSON_string> <user_auth_token>",
		Short: "Update user",
		Long:  `Update user metadata`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			var user mfxsdk.User
			if err := json.Unmarshal([]byte(args[0]), &user.Metadata); err != nil {
				logError(err)
				return
			}

			if err := sdk.UpdateUser(user, args[1]); err != nil {
				logError(err)
				return
			}

			logOK()
		},
	},
	{
		Use:   "password <old_password> <password> <user_auth_token>",
		Short: "Update password",
		Long:  `Update user password`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Use)
				return
			}

			if err := sdk.UpdatePassword(args[0], args[1], args[2]); err != nil {
				logError(err)
				return
			}

			logOK()
		},
	},
}

// NewUsersCmd returns users command.
func NewUsersCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "users [create | get | update | token | password]",
		Short: "Users management",
		Long:  `Users management: create accounts and tokens"`,
	}

	for i := range cmdUsers {
		cmd.AddCommand(&cmdUsers[i])
	}

	return &cmd
}
