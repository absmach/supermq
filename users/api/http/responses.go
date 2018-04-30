package http

import "net/http"

type apiRes interface {
	code() int
	headers() map[string]string
	empty() bool
}

type tokenRes struct {
	Token string `json:"token,omitempty"`
}

func (res tokenRes) code() int {
	return http.StatusCreated
}

func (res tokenRes) headers() map[string]string {
	return map[string]string{}
}

func (res tokenRes) empty() bool {
	return res.Token == ""
}
