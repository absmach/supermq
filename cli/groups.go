// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"

	mfclients "github.com/mainflux/mainflux/pkg/clients"
	mfxsdk "github.com/mainflux/mainflux/pkg/sdk/go"
	"github.com/spf13/cobra"
)

var cmdGroups = []cobra.Command{
	{
		Use:   "create <JSON_group> <user_auth_token>",
		Short: "Create group",
		Long: `Creates new group:
		{
			"Name":<group_name>,
			"Description":<description>,
			"ParentID":<parent_id>,
			"Metadata":<metadata>,
		}
		Name - is unique group name
		ParentID - ID of a group that is a parent to the creating group
		Metadata - JSON structured string`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}
			var group mfxsdk.Group
			if err := json.Unmarshal([]byte(args[0]), &group); err != nil {
				logError(err)
				return
			}
			group.Status = mfclients.EnabledStatus.String()
			group, err := sdk.CreateGroup(group, args[1])
			if err != nil {
				logError(err)
				return
			}
			logJSON(group)
		},
	},
	{
		Use:   "get [all | children <group_id> | parents <group_id> | <group_id>] <user_auth_token>",
		Short: "Get group",
		Long: `Get all users groups, group children or group by id.
		all - lists all groups
		children <group_id> - lists all children groups of <group_id>
		<group_id> - shows group with provided group ID`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 2 {
				logUsage(cmd.Use)
				return
			}
			if args[0] == all {
				if len(args) > 2 {
					logUsage(cmd.Use)
					return
				}
				pm := mfxsdk.PageMetadata{
					Offset: uint64(Offset),
					Limit:  uint64(Limit),
				}
				l, err := sdk.Groups(pm, args[1])
				if err != nil {
					logError(err)
					return
				}
				logJSON(l)
				return
			}
			if args[0] == "children" {
				if len(args) > 3 {
					logUsage(cmd.Use)
					return
				}
				pm := mfxsdk.PageMetadata{
					Offset: uint64(Offset),
					Limit:  uint64(Limit),
				}
				l, err := sdk.Children(args[1], pm, args[2])
				if err != nil {
					logError(err)
					return
				}
				logJSON(l)
				return
			}
			if args[0] == "parents" {
				if len(args) > 3 {
					logUsage(cmd.Use)
					return
				}
				pm := mfxsdk.PageMetadata{
					Offset: uint64(Offset),
					Limit:  uint64(Limit),
				}
				l, err := sdk.Parents(args[1], pm, args[2])
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
			t, err := sdk.Group(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}
			logJSON(t)
		},
	},
	{
		Use:   "assign <member_type> <member_id> <group_id> <user_auth_token>",
		Short: "Assign member",
		Long: `Assign members to a group.
				member_ids - '["member_id",...]`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 4 {
				logUsage(cmd.Use)
				return
			}
			var types []string
			if err := json.Unmarshal([]byte(args[0]), &types); err != nil {
				logError(err)
				return
			}
			if err := sdk.Assign(types, args[1], args[2], args[3]); err != nil {
				logError(err)
				return
			}
			logOK()
		},
	},
	{
		Use:   "unassign <member_type> <group_id> <member_id> <user_auth_token>",
		Short: "Unassign member",
		Long: `Unassign members from a group
				member_ids - '["member_id",...]`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 4 {
				logUsage(cmd.Use)
				return
			}
			var types []string
			if err := json.Unmarshal([]byte(args[0]), &types); err != nil {
				logError(err)
				return
			}
			if err := sdk.Unassign(types, args[1], args[2], args[3]); err != nil {
				logError(err)
				return
			}
			logOK()
		},
	},
	{
		Use:   "members <group_id> <user_auth_token>",
		Short: "Members list",
		Long:  `Lists all members of a group.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}
			pm := mfxsdk.PageMetadata{
				Offset: uint64(Offset),
				Limit:  uint64(Limit),
				Status: Status,
			}
			up, err := sdk.Members(args[0], pm, args[1])
			if err != nil {
				logError(err)
				return
			}
			logJSON(up)
		},
	},
	{
		Use:   "membership <member_id> <user_auth_token>",
		Short: "Membership list",
		Long:  `List member group's membership`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}
			pm := mfxsdk.PageMetadata{
				Offset: uint64(Offset),
				Limit:  uint64(Limit),
			}
			up, err := sdk.Memberships(args[0], pm, args[1])
			if err != nil {
				logError(err)
				return
			}
			logJSON(up)
		},
	},
	{
		Use:   "enable <group_id> <user_auth_token>",
		Short: "Change group status to enabled",
		Long:  `Change group status to enabled`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			group, err := sdk.EnableGroup(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(group)
		},
	},
	{
		Use:   "disable <group_id> <user_auth_token>",
		Short: "Change group status to disabled",
		Long:  `Change group status to disabled`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Use)
				return
			}

			group, err := sdk.DisableGroup(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(group)
		},
	},
}

// NewGroupsCmd returns users command.
func NewGroupsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "groups [create | get | delete | assign | unassign | members | membership]",
		Short: "Groups management",
		Long:  `Groups management: create groups and assigns member to groups"`,
	}

	for i := range cmdGroups {
		cmd.AddCommand(&cmdGroups[i])
	}

	return &cmd
}
