package cmd

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fatih/color"
	"github.com/hokaccha/go-prettyjson"
)

const contentType = "application/json"

var Limit = 10
var Offset = 0

func GetReqResp(req *http.Request, token string) {
	req.Header.Set("Authorization", token)
	req.Header.Add("Content-Type", contentType)

	resp, err := httpClient.Do(req)
	FormatResLog(resp, err)
}

// FormatResLog - format http response
func FormatResLog(resp *http.Response, err error) {
	if err != nil {
		fmt.Println(err.Error() + "\n")
	}
	defer resp.Body.Close()

	log := color.YellowString(fmt.Sprintf("%s %s\n%s %v\n",
		resp.Proto, resp.Status, "Content-Length: ", resp.ContentLength))
	fmt.Println(log)

	if len(resp.Header.Get("Location")) != 0 {
		log = fmt.Sprintf("%s %s\n", "Resource location:", resp.Header.Get("Location"))
		fmt.Println(log)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err.Error() + "\n")
		return
	}

	if len(body) != 0 {
		pj, err := prettyjson.Format([]byte(body))
		if err != nil {
			fmt.Println(string(body) + "\n")
			return
		}
		fmt.Println(string(pj) + "\n")
	}
}

func LogUsage(u string) {
	fmt.Println("Usage: ", u)
}
