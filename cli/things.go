//
// Copyright (c) 2018
// Mainflux
//
// SPDX-License-Identifier: Apache-2.0
//

package cli

import (
	"github.com/mainflux/mainflux/sdk/go"
	"github.com/spf13/cobra"
)

const thingsEP = "things"

var cmdThings = []cobra.Command{
	cobra.Command{
		Use:   "create",
		Short: "create <JSON_thing> <user_auth_token>",
		Long:  `Create new thing, generate his UUID and store it`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}
			id, err := sdk.CreateThing(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}
			dump(id)
		},
	},
	cobra.Command{
		Use:   "get",
		Short: "get all/<thing_id> <user_auth_token>",
		Long:  `Get all things or thing by id`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}
			if args[0] == "all" {
				l, err := sdk.GetThings(args[1])
				if err != nil {
					logError(err)
					return
				}
				dump(l)
				return
			}
			t, err := sdk.GetThing(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}
			dump(t)
		},
	},
	cobra.Command{
		Use:   "delete",
		Short: "delete <thing_id> <user_auth_token>",
		Long:  `Removes thing from database`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}
			if err := sdk.DeleteThing(args[0], args[1]); err != nil {
				logError(err)
				return
			}
			logOK()
		},
	},
	cobra.Command{
		Use:   "update",
		Short: "update <thing_id> <JSON_string> <user_auth_token>",
		Long:  `Update thing record`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Short)
				return
			}
			if err := sdk.UpdateThing(args[0], args[1], args[2]); err != nil {
				logError(err)
				return
			}
			logOK()
		},
	},
	cobra.Command{
		Use:   "connect",
		Short: "connect <thing_id> <channel_id> <user_auth_token>",
		Long:  `Connect thing to the channel`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Short)
				return
			}
			if err := sdk.ConnectThing(args[0], args[1], args[2]); err != nil {
				logError(err)
				return
			}
			logOK()
		},
	},
	cobra.Command{
		Use:   "disconnect",
		Short: "disconnect <thing_id> <channel_id> <user_auth_token>",
		Long:  `Disconnect thing to the channel`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Short)
				return
			}
			if err := sdk.DisconnectThing(args[0], args[1], args[2]); err != nil {
				logError(err)
				return
			}
			logOK()
		},
	},
}

func NewThingsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "things",
		Short: "things <options>",
		Long:  `Things handling: create, delete or update things`,
		Run: func(cmd *cobra.Command, args []string) {
			logUsage(cmd.Short)
		},
	}

	for i := range cmdThings {
		cmd.AddCommand(&cmdThings[i])
	}

	return &cmd
}
