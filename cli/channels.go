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

var cmdChannels = []cobra.Command{
	cobra.Command{
		Use:   "create",
		Short: "create <JSON_channel> <user_auth_token>",
		Long:  `Creates new channel and generates it's UUID`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				LogUsage(cmd.Short)
				return
			}
			FormatResLog(sdk.CreateChannel(args[0], args[1]))
		},
	},
	cobra.Command{
		Use:   "get",
		Short: "get all/<channel_id> <user_auth_token>",
		Long:  `Gets list of all channels or gets channel by id`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				LogUsage(cmd.Short)
				return
			}
			if args[0] == "all" {
				FormatResLog(sdk.GetChannels(args[1]))
				return
			}
			FormatResLog(sdk.GetChannel(args[0], args[1]))
		},
	},
	cobra.Command{
		Use:   "update",
		Short: "update <channel_id> <JSON_string> <user_auth_token>",
		Long:  `Updates channel record`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				LogUsage(cmd.Short)
				return
			}
			FormatResLog(sdk.UpdateChannel(args[0], args[1], args[2]))
		},
	},
	cobra.Command{
		Use:   "delete",
		Short: "delete <channel_id> <user_auth_token>",
		Long:  `Delete channel by ID`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				LogUsage(cmd.Short)
				return
			}
			FormatResLog(sdk.DeleteChannel(args[0], args[1]))
		},
	},
}

func NewChannelsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "channels",
		Short: "Manipulation with channels",
		Long:  `Manipulation with channels: create, delete or update channels`,
		Run: func(cmd *cobra.Command, args []string) {
			LogUsage(cmd.Short)
		},
	}

	for i, _ := range cmdChannels {
		cmd.AddCommand(&cmdChannels[i])
	}

	return &cmd
}
