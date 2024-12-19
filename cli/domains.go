// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"

	smqsdk "github.com/absmach/supermq/pkg/sdk"
	"github.com/spf13/cobra"
)

var cmdDomains = []cobra.Command{
	{
		Use:   "create <name> <alias> <token>",
		Short: "Create Domain",
		Long: "Create Domain with provided name and alias. \n" +
			"For example:\n" +
			"\tsupermq-cli domains create domain_1 domain_1_alias $TOKEN\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 3 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			dom := smqsdk.Domain{
				Name:  args[0],
				Alias: args[1],
			}
			d, err := sdk.CreateDomain(dom, args[2])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			logJSONCmd(*cmd, d)
		},
	},
	{
		Use:   "get [all | <domain_id> ] <token>",
		Short: "Get Domains",
		Long:  "Get all domains. Users can be filtered by name or metadata or status",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}
			metadata, err := convertMetadata(Metadata)
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			pageMetadata := smqsdk.PageMetadata{
				Name:     Name,
				Offset:   Offset,
				Limit:    Limit,
				Metadata: metadata,
				Status:   Status,
			}
			if args[0] == all {
				l, err := sdk.Domains(pageMetadata, args[1])
				if err != nil {
					logErrorCmd(*cmd, err)
					return
				}
				logJSONCmd(*cmd, l)
				return
			}
			d, err := sdk.Domain(args[0], args[1])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}

			logJSONCmd(*cmd, d)
		},
	},

	{
		Use:   "users <domain_id>  <token>",
		Short: "List Domain users",
		Long:  "List Domain users",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}
			metadata, err := convertMetadata(Metadata)
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			pageMetadata := smqsdk.PageMetadata{
				Offset:   Offset,
				Limit:    Limit,
				Metadata: metadata,
				Status:   Status,
			}

			l, err := sdk.ListDomainUsers(args[0], pageMetadata, args[1])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			logJSONCmd(*cmd, l)
		},
	},

	{
		Use:   "update <domain_id> <JSON_string> <user_auth_token>",
		Short: "Update domains",
		Long: "Updates domains name, alias and metadata \n" +
			"Usage:\n" +
			"\tsupermq-cli domains update <domain_id> '{\"name\":\"new name\", \"alias\":\"new_alias\", \"metadata\":{\"key\": \"value\"}}' $TOKEN \n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 4 && len(args) != 3 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			var d smqsdk.Domain

			if err := json.Unmarshal([]byte(args[1]), &d); err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			d.ID = args[0]
			d, err := sdk.UpdateDomain(d, args[2])
			if err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			logJSONCmd(*cmd, d)
		},
	},

	{
		Use:   "enable <domain_id> <token>",
		Short: "Change domain status to enabled",
		Long: "Change domain status to enabled\n" +
			"Usage:\n" +
			"\tsupermq-cli domains enable <domain_id> <token>\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			if err := sdk.EnableDomain(args[0], args[1]); err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			logOKCmd(*cmd)
		},
	},
	{
		Use:   "disable <domain_id> <token>",
		Short: "Change domain status to disabled",
		Long: "Change domain status to disabled\n" +
			"Usage:\n" +
			"\tsupermq-cli domains disable <domain_id> <token>\n",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsageCmd(*cmd, cmd.Use)
				return
			}

			if err := sdk.DisableDomain(args[0], args[1]); err != nil {
				logErrorCmd(*cmd, err)
				return
			}
			logOKCmd(*cmd)
		},
	},
}

// NewDomainsCmd returns domains command.
func NewDomainsCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "domains [create | get | update | enable | disable | enable | users | assign | unassign]",
		Short: "Domains management",
		Long:  `Domains management: create, update, retrieve domains , assign/unassign users to domains and list users of domain"`,
	}

	for i := range cmdDomains {
		cmd.AddCommand(&cmdDomains[i])
	}

	return &cmd
}
