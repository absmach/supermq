// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"

	"github.com/absmach/supermq/groups"
	smqsdk "github.com/absmach/supermq/pkg/sdk"
	"github.com/spf13/cobra"
)

var cmdGroups = []cobra.Command{
	{
		Use:   "create <JSON_group> <domain_id> <user_auth_token>",
		Short: "Create group",
		Long: "Creates new group\n" +
			"Usage:\n" +
			"\tsupermq-cli groups create '{\"name\":\"new group\", \"description\":\"new group description\", \"metadata\":{\"key\": \"value\"}}' $DOMAINID $USERTOKEN\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}
			var group smqsdk.Group
			if err := json.Unmarshal([]byte(args[0]), &group); err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			group.Status = groups.EnabledStatus.String()
			group, err := sdk.CreateGroup(group, args[1], args[2])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			logJSONCmd(*cmd, group)
		},
	},
	{
		Use:   "update <JSON_group> <domain_id> <user_auth_token>",
		Short: "Update group",
		Long: "Updates group\n" +
			"Usage:\n" +
			"\tsupermq-cli groups update '{\"id\":\"<group_id>\", \"name\":\"new group\", \"description\":\"new group description\", \"metadata\":{\"key\": \"value\"}}' $DOMAINID $USERTOKEN\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			var group smqsdk.Group
			if err := json.Unmarshal([]byte(args[0]), &group); err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			group, err := sdk.UpdateGroup(group, args[1], args[2])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logJSONCmd(*cmd, group)
		},
	},
	{
		Use:   "delete <group_id> <domain_id> <user_auth_token>",
		Short: "Delete group",
		Long: "Delete group by id.\n" +
			"Usage:\n" +
			"\tsupermq-cli groups delete <group_id> $DOMAINID $USERTOKEN - delete the given group ID\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}
			if err := sdk.DeleteGroup(args[0], args[1], args[2]); err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			logOKCmd(*cmd)
		},
	},
	{
		Use:   "enable <group_id> <domain_id> <user_auth_token>",
		Short: "Change group status to enabled",
		Long: "Change group status to enabled\n" +
			"Usage:\n" +
			"\tsupermq-cli groups enable <group_id> $DOMAINID $USERTOKEN\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			group, err := sdk.EnableGroup(args[0], args[1], args[2])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logJSONCmd(*cmd, group)
		},
	},
	{
		Use:   "disable <group_id> <domain_id> <user_auth_token>",
		Short: "Change group status to disabled",
		Long: "Change group status to disabled\n" +
			"Usage:\n" +
			"\tsupermq-cli groups disable <group_id> $DOMAINID $USERTOKEN\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			group, err := sdk.DisableGroup(args[0], args[1], args[2])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logJSONCmd(*cmd, group)
		},
	},
}

// NewGroupsCmd returns users command.
func NewGroupsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "groups [create | get | update | delete | assign | unassign | users | channels ]",
		Short: "Groups management",
		Long:  `Groups management: create, update, delete group and assign and unassign member to groups"`,
	}

	for i := range cmdGroups {
		cmd.AddCommand(&cmdGroups[i])
	}

	return &cmd
}
