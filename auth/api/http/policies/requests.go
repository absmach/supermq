package policies

import (
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/things"
)

const (
	readPolicy   = "read"
	writePolicy  = "write"
	deletePolicy = "delete"
	accessPolicy = "access"
	memberPolicy = "member"
	createPolicy = "create"
)

type createPolicyReq struct {
	token      string
	SubjectIDs []string `json:"subjects"`
	Policies   []string `json:"policies"`
	Object     string   `json:"object"`
}

func (req createPolicyReq) validate() error {
	if req.token == "" {
		return auth.ErrUnauthorizedAccess
	}

	if len(req.SubjectIDs) == 0 || len(req.Policies) == 0 || req.Object == "" {
		return auth.ErrMalformedEntity
	}

	for _, policy := range req.Policies {
		if policy != readPolicy && policy != writePolicy && policy != deletePolicy &&
			policy != accessPolicy && policy != memberPolicy && policy != createPolicy {
			return things.ErrMalformedEntity
		}
	}

	for _, subject := range req.SubjectIDs {
		if subject == "" {
			return things.ErrMalformedEntity
		}
	}

	return nil
}
