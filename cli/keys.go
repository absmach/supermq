package cli

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

var cmdAPIKeys = []cobra.Command{
	cobra.Command{
		Use:   "issue",
		Short: "issue <api_key_token>",
		Long:  `Issues a new Key`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}

			var apikey mfxsdk.issueKeyReq
			if err := json.Unmarshal([]byte(args[0]), &apikey); err != nil {
				logError(err)
				return
			}

			id, err := sdk.Issue(apikey, args[1])
			if err != nil {
				logError(err)
				return
			}

			logCreated(id)
		},
	},
	cobra.Command{
		Use:   "revoke",
		Short: "revoke <api_key_token> <key_id>",
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
	cobra.Command{
		Use:   "RetrieveKey",
		Short: "retrievekey <key_id> <api_key_token>",
		Long:  `Get API key id`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) != 2 {
				logUsage(cmd.Short)
				return
			}

			ak, err := sdk.RetrieveKey(args[0], args[1])
			if err != nil {
				logError(err)
				return
			}

			logJSON(ak)
		},
	},
}
