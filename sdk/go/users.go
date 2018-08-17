package sdk

import (
	"fmt"
	"net/http"
	"strings"
)

// CreateUser - create user
func CreateUser(user, pwd string) (*http.Response, error) {
	msg := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, user, pwd)
	url := fmt.Sprintf("%s/users", serverAddr)

	return httpClient.Post(url, contentTypeJSON, strings.NewReader(msg))
}

// CreateToken - create user token
func CreateToken(user, pwd string) (*http.Response, error) {
	msg := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, user, pwd)
	url := fmt.Sprintf("%s/tokens", serverAddr)

	return httpClient.Post(url, contentTypeJSON, strings.NewReader(msg))
}
