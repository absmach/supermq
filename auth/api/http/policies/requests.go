package policies

import "github.com/mainflux/mainflux/auth"

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

	return nil
}
