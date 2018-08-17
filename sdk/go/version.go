package sdk

import (
	"fmt"
	"net/http"
)

// Version - server health check
func Version() (*http.Response, error) {
	url := fmt.Sprintf("%s/version", serverAddr)

	return httpClient.Get(url)
}
