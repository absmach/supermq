package policies

import "net/http"

type createPolicyRes struct{}

func (res createPolicyRes) Code() int {
	return http.StatusOK
}

func (res createPolicyRes) Headers() map[string]string {
	return map[string]string{}
}

func (res createPolicyRes) Empty() bool {
	return false
}

type errorRes struct {
	Err string `json:"error"`
}
