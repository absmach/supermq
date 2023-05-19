package policies

import (
	"context"
	"time"

	"github.com/mainflux/mainflux/internal/apiutil"
)

// PolicyTypes contains a list of the available policy types currently supported
var PolicyTypes = []string{WriteAction, ReadAction}

// Policy represents an argument struct for making a policy related function calls.
type Policy struct {
	OwnerID   string    `json:"owner_id"`
	Subject   string    `json:"subject"`
	Object    string    `json:"object"`
	Actions   []string  `json:"actions"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
}

// AccessRequest represents an access control request for Authorization.
type AccessRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

// PolicyPage contains a page of policies.
type PolicyPage struct {
	Page
	Policies []Policy
}

// Repository specifies an account persistence API.
type Repository interface {
	// Save creates a policy for the given Subject, so that, after
	// Save, `Subject` has a `relation` on `group_id`. Returns a non-nil
	// error in case of failures.
	Save(ctx context.Context, p Policy) (Policy, error)

	// Evaluate is used to evaluate if you have the correct permissions.
	// We evaluate if we are in the same group first then evaluate if the
	// object has that action over the subject
	Evaluate(ctx context.Context, entityType string, p Policy) error

	// RetrieveOne retrieves policy by subject and object.
	RetrieveOne(ctx context.Context, subject, object string) (Policy, error)

	// Update updates the policy type.
	Update(ctx context.Context, p Policy) (Policy, error)

	// Retrieve retrieves policy for a given input.
	Retrieve(ctx context.Context, pm Page) (PolicyPage, error)

	// Delete deletes the policy
	Delete(ctx context.Context, p Policy) error
}

// Service represents a authorization service. It exposes
// functionalities through `auth` to perform authorization.
type Service interface {
	// Authorize checks authorization of the given `subject`.
	// Authorize verifies that Is `subject` allowed to `relation` on
	// `object`. Authorize returns a non-nil error if the subject has
	// no relation on the object (which simply means the operation is
	// denied).
	Authorize(ctx context.Context, ar AccessRequest, entity string) (string, error)

	// AddPolicy creates a policy for the given subject, so that, after
	// AddPolicy, `subject` has a `relation` on `object`. Returns a non-nil
	// error in case of failures.
	AddPolicy(ctx context.Context, token string, p Policy) (Policy, error)

	// DeletePolicy removes a policy.
	DeletePolicy(ctx context.Context, token string, p Policy) error

	// UpdatePolicy updates an existing policy
	UpdatePolicy(ctx context.Context, token string, p Policy) (Policy, error)

	// ListPolicies lists existing policies
	ListPolicies(ctx context.Context, token string, p Page) (PolicyPage, error)
}

// Cache contains channel-thing connection caching interface.
type Cache interface {
	// Put connects group to a client with the specified action.
	Put(ctx context.Context, policy Policy) error

	// Get checks if a client is connected to group.
	Get(ctx context.Context, policy Policy) (Policy, error)

	// Remove deletes a client connection to a group.
	Remove(ctx context.Context, policy Policy) error
}

// ClientCache contains thing caching interface.
type ClientCache interface {
	// Save stores pair thing key, thing id.
	Save(context.Context, string, string) error

	// ID returns thing ID for given key.
	ID(context.Context, string) (string, error)

	// Removes thing from cache.
	Remove(context.Context, string) error
}

// Validate returns an error if policy representation is invalid.
func (p Policy) Validate() error {
	if p.Subject == "" {
		return apiutil.ErrMissingPolicySub
	}
	if p.Object == "" {
		return apiutil.ErrMissingPolicyObj
	}
	if len(p.Actions) == 0 {
		return apiutil.ErrMalformedPolicyAct
	}
	for _, p := range p.Actions {
		// Validate things policies first
		if ok := ValidateAction(p); !ok {
			// Validate users policies for clients connected to a group
			if ok := ValidateAction(p); !ok {
				return apiutil.ErrMalformedPolicyAct
			}
		}
	}
	return nil
}

// ValidateAction check if the action is in policies
func ValidateAction(act string) bool {
	for _, v := range PolicyTypes {
		if v == act {
			return true
		}
	}
	return false

}
