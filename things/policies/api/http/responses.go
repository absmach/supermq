package api

import (
	"fmt"
	"net/http"

	"github.com/mainflux/mainflux"
)

var (
	_ mainflux.Response = (*addPolicyRes)(nil)
	_ mainflux.Response = (*identityRes)(nil)
	_ mainflux.Response = (*canAccessByIDRes)(nil)
	_ mainflux.Response = (*deletePolicyRes)(nil)
)

type addPolicyRes struct {
	id      string
	created bool
}

func (res addPolicyRes) Code() int {
	if res.created {
		return http.StatusCreated
	}

	return http.StatusOK
}

func (res addPolicyRes) Headers() map[string]string {
	if res.created {
		return map[string]string{
			"Location": fmt.Sprintf("/groups/%s", res.id),
		}
	}

	return map[string]string{}
}

func (res addPolicyRes) Empty() bool {
	return true
}

type deletePolicyRes struct{}

func (res deletePolicyRes) Code() int {
	return http.StatusNoContent
}

func (res deletePolicyRes) Headers() map[string]string {
	return map[string]string{}
}

func (res deletePolicyRes) Empty() bool {
	return true
}

type identityRes struct {
	ID string `json:"id"`
}

func (res identityRes) Code() int {
	return http.StatusOK
}

func (res identityRes) Headers() map[string]string {
	return map[string]string{}
}

func (res identityRes) Empty() bool {
	return false
}

type canAccessByIDRes struct{}

func (res canAccessByIDRes) Code() int {
	return http.StatusOK
}

func (res canAccessByIDRes) Headers() map[string]string {
	return map[string]string{}
}

func (res canAccessByIDRes) Empty() bool {
	return true
}
