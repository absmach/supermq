// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"

	mfxsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/spf13/cobra"
)

var cmdBootstrap = []cobra.Command{
	{
		Use:   "add {JSON_config} {user_auth_token}",
		Short: "Add config",
		Long:  `Adds new Thing Bootstrap Config to the user identified by the provided key`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			var cfg mfxsdk.BootstrapConfig
			if err := json.Unmarshal([]byte(args[0]), &cfg); err != nil {
				logError(err)
				return
			}

			id, err := sdk.AddBootstrap(args[1], cfg)
			if err != nil {
				logError(err)
				return
			}

			logCreated(id)
		},
	},
	{
		Use:   "view {thing_id} {user_auth_token}",
		Short: "View config",
		Long:  `Returns Thing Config with given ID belonging to the user identified by the given key`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			c, err := sdk.ViewBootstrap(args[1], args[0])
			if err != nil {
				logError(err)
				return
			}

			logJSON(c)
		},
	},
	{
		Use:   "update {JSON_config} {user_auth_token}",
		Short: "Update config",
		Long:  `Updates editable fields of the provided Config`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			var cfg mfxsdk.BootstrapConfig
			if err := json.Unmarshal([]byte(args[0]), &cfg); err != nil {
				logError(err)
				return
			}

			if err := sdk.UpdateBootstrap(args[1], cfg); err != nil {
				logError(err)
				return
			}

			logOK()
		},
	},
	{
		Use:   "remove {thing_id} {user_auth_token}",
		Short: "Remove config",
		Long:  `Removes Config with specified key that belongs to the user identified by the given key`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			if err := sdk.RemoveBootstrap(args[1], args[0]); err != nil {
				logError(err)
				return
			}

			logOK()
		},
	},
	{
		Use:   "bootstrap {external_id} {external_key}",
		Short: "Bootstrap config",
		Long:  `Returns Config to the Thing with provided external ID using external key`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			c, err := sdk.Bootstrap(args[1], args[0])
			if err != nil {
				logError(err)
				return
			}

			logJSON(c)
		},
	},
}

// NewBootstrapCmd returns bootstrap command.
func NewBootstrapCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "bootstrap [add | view | update | remove | bootstrap]",
		Short: "Bootstrap management",
		Long:  `Bootstrap management: create, get, update or delete Bootstrap config`,
	}

	for i := range cmdBootstrap {
		cmd.AddCommand(&cmdBootstrap[i])
	}

	return &cmd
}
