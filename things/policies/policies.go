package policies

import (
	"context"
	"time"

	"github.com/mainflux/mainflux/internal/apiutil"
)

// PolicyTypes contains a list of the available policy types currently supported
var PolicyTypes = []string{"c_delete", "c_update", "c_add", "c_list", "g_delete", "g_update", "g_add", "g_list", "m_write", "m_read"}

// Policy represents an argument struct for making a policy related function calls.
type Policy struct {
	OwnerID   string    `json:"owner_id"`
	Subject   string    `json:"subject"`
	Object    string    `json:"object"`
	Actions   []string  `json:"actions"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PolicyPage contains a page of policies.
type PolicyPage struct {
	Page
	Policies []Policy
}

// PolicyRepository specifies an account persistence API.
type PolicyRepository interface {
	// Save creates a policy for the given Subject, so that, after
	// Save, `Subject` has a `relation` on `group_id`. Returns a non-nil
	// error in case of failures.
	Save(ctx context.Context, p Policy) error

	// Evaluate is used to evaluate if you have the correct permissions.
	// We evaluate if we are in the same group first then evaluate if the
	// object has that action over the subject
	Evaluate(ctx context.Context, entityType string, p Policy) error

	// Update updates the policy type.
	Update(ctx context.Context, p Policy) error

	// Retrieve retrieves policy for a given input.
	Retrieve(ctx context.Context, pm Page) (PolicyPage, error)

	// Delete deletes the policy
	Delete(ctx context.Context, p Policy) error

	HasThingByID(ctx context.Context, chanID, thingID string) error

	HasThing(ctx context.Context, chanID, thingKey string) (string, error)
}

// PolicyService represents a authorization service. It exposes
// functionalities through `auth` to perform authorization.
type PolicyService interface {
	// AddPolicy creates a policy for the given subject, so that, after
	// AddPolicy, `subject` has a `relation` on `object`. Returns a non-nil
	// error in case of failures.
	AddPolicy(ctx context.Context, token string, p Policy) error

	// DeletePolicy removes a policy.
	DeletePolicy(ctx context.Context, token string, p Policy) error

	// CanAccessByKey determines whether the channel can be accessed using the
	// provided key and returns thing's id if access is allowed.
	CanAccessByKey(ctx context.Context, chanID, key string) (string, error)

	// CanAccessByID determines whether the channel can be accessed by
	// the given thing and returns error if it cannot.
	CanAccessByID(ctx context.Context, chanID, thingID string) error
}

// ChannelCache contains channel-thing connection caching interface.
type ChannelCache interface {
	// Connect channel thing connection.
	Connect(context.Context, string, string) error

	// HasThing checks if thing is connected to channel.
	HasThing(context.Context, string, string) bool

	// Disconnects thing from channel.
	Disconnect(context.Context, string, string) error

	// Removes channel from cache.
	Remove(context.Context, string) error
}

// ThingCache contains thing caching interface.
type ThingCache interface {
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
		return apiutil.ErrMissingPolicyAct
	}
	for _, p := range p.Actions {
		if ok := ValidateAction(p); !ok {
			return apiutil.ErrMissingPolicyAct
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
