package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

var CmdMessages = &cobra.Command{
	Use:   "msg",
	Short: "Send or retrieve messages",
	Long:  `Send or retrieve messages: control message flow on the channel`,
}

var CmdSendMessage = &cobra.Command{
	Use:   "send",
	Short: "send <channel_id> <JSON_string> <client_token>",
	Long:  `Sends message on the channel`,
	Run: func(cmdCobra *cobra.Command, args []string) {
		if len(args) == 3 {
			SendMsg(args[0], args[1], args[2])
		} else {
			LogUsage(cmdCobra.Short)
		}
	},
}

// SendMsg - publishes SenML message on the channel
func SendMsg(id string, msg string, token string) {
	url := serverAddr + "/channels/" + id + "/messages"
	req, err := http.NewRequest("POST", url, strings.NewReader(msg))
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}

	req.Header.Set("Authorization", token)
	req.Header.Add("Content-Type", "application/senml+json")

	resp, err := httpClient.Do(req)
	FormatResLog(resp, err)
}
