package policies

import (
	"github.com/mainflux/mainflux/auth"
	"github.com/mainflux/mainflux/things"
)

// Action represents an enum for the policies used in the Mainflux.
type Action int

const (
	Create Action = iota
	Read
	Write
	Delete
	Access
	Member
	unknown
)

var actions = [...]string{
	Create: "create",
	Read:   "read",
	Write:  "write",
	Delete: "delete",
	Access: "access",
	Member: "member",
}

func parsePolicy(incomingAction string) Action {
	for i, action := range actions {
		if incomingAction == action {
			return Action(i)
		}
	}
	return unknown
}

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
		if action := parsePolicy(policy); action > Member || action < Create {
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
