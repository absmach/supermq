// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"strconv"
	"time"

	mfxsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/spf13/cobra"
)

var cmdAPIKeys = []cobra.Command{
	{
		Use:   "issue",
		Short: "issue type duration <api_key_token>",
		Long:  `Issues a new Key`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsage(cmd.Short)
				return
			}

			var req mfxsdk.KeyReq
			t, err := strconv.Atoi(args[0])
			if err != nil {
				logError(err)
				return
			}
			req.Type = uint32(t)
			d, err := time.ParseDuration(args[1])
			if err != nil {
				logError(err)
				return
			}
			req.Duration = d
			resp, err := sdk.Issue(args[2], req)
			if err != nil {
				logError(err)
				return
			}

			logCreated(resp.ID)
		},
	},
	{
		Use:   "revoke",
		Short: "revoke <key_id> <user_auth_token>",
		Long:  `Removes API key from database`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}

			if err := sdk.Revoke(args[0], args[1]); err != nil {
				logError(err)
				return
			}

			logOK()
		},
	},
	{
		Use:   "retrieve",
		Short: "retrieve <key_id> <user_auth_token>",
		Long:  `Retrieve API key with given id`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}

			rk, err := sdk.RetrieveKey(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(rk)
		},
	},
}

// NewKeysCmd returns keys command.
func NewKeysCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "keys",
		Short: "Keys management",
		Long:  `Keys management: create, get, update or delete Thing, connect or disconnect Thing from Channel and get the list of Channels connected or disconnected from a Thing`,
		Run: func(cmd *cobra.Command, args []string) {
			logUsage("keys [issue | revoke | retrive]")
		},
	}

	for i := range cmdAPIKeys {
		cmd.AddCommand(&cmdAPIKeys[i])
	}

	return &cmd
}
