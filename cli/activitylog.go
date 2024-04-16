// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	mgxsdk "github.com/absmach/magistrala/pkg/sdk/go"
	"github.com/spf13/cobra"
)

var cmdActivity = cobra.Command{
	Use:   "get <user_auth_token>",
	Short: "Get activities",
	Long: "Get activities\n" +
		"Usage:\n" +
		"\tmagistrala-cli activities get <user_auth_token> - lists all activities\n" +
		"\tmagistrala-cli activities get <user_auth_token> --offset <offset> --limit <limit> - lists all activities with provided offset and limit\n",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			logUsage(cmd.Use)
			return
		}

		pageMetadata := mgxsdk.PageMetadata{
			Offset: Offset,
			Limit:  Limit,
		}

		activities, err := sdk.Activities(pageMetadata, args[0])
		if err != nil {
			logError(err)
			return
		}

		logJSON(activities)
	},
}

// NewActivitiesCmd returns activity log command.
func NewActivitiesCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "activities get",
		Short: "activity log",
		Long:  `activities to read activity log`,
	}

	cmd.AddCommand(&cmdActivity)

	return &cmd
}
