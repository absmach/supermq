package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mainflux/mainflux/pkg/errors"
	"golang.org/x/net/idna"
)

const (
	maxLocalLen  = 64
	maxDomainLen = 255
	maxTLDLen    = 24 // longest TLD currently in existence

	atSeparator  = "@"
	dotSeparator = "."
)

var (
	userRegexp    = regexp.MustCompile("^[a-zA-Z0-9!#$%&'*+/=?^_`{|}~.-]+$")
	hostRegexp    = regexp.MustCompile(`^[^\s]+\.[^\s]+$`)
	userDotRegexp = regexp.MustCompile("(^[.]{1})|([.]{1}$)|([.]{2,})")
)

// Credentials represent client credentials: its
// "identity" which can be a username, email, generated name;
// and "secret" which can be a password or access token.
type Credentials struct {
	Identity string `json:"identity"` // username or generated login ID
	Secret   string `json:"secret"`   // password or token
}

// Client represents generic Client.
type Client struct {
	ID          string      `json:"id"`
	Name        string      `json:"name,omitempty"`
	Tags        []string    `json:"tags,omitempty"`
	Owner       string      `json:"owner,omitempty"` // nullable
	Credentials Credentials `json:"credentials"`
	Metadata    Metadata    `json:"metadata,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Status      Status      `json:"status"`         // 1 for enabled, 0 for disabled
	Role        Role        `json:"role,omitempty"` // 1 for admin, 0 for normal user
}

type UserIdentity struct {
	ID    string
	Email string
}

// ClientsPage contains page related metadata as well as list
// of Clients that belong to the page.
type ClientsPage struct {
	Page
	Clients []Client
}

// MembersPage contains page related metadata as well as list of members that
// belong to this page.
type MembersPage struct {
	Page
	Members []Client
}

// ClientRepository specifies an account persistence API.
type ClientRepository interface {
	// Save persists the client account. A non-nil error is returned to indicate
	// operation failure.
	Save(ctx context.Context, client Client) (Client, error)

	// RetrieveByID retrieves client by its unique ID.
	RetrieveByID(ctx context.Context, id string) (Client, error)

	// RetrieveByIdentity retrieves client by its unique credentials
	RetrieveByIdentity(ctx context.Context, identity string) (Client, error)

	// RetrieveAll retrieves all clients.
	RetrieveAll(ctx context.Context, pm Page) (ClientsPage, error)

	// Members retrieves everything that is assigned to a group identified by groupID.
	Members(ctx context.Context, groupID string, pm Page) (MembersPage, error)

	// Update updates the client name and metadata.
	Update(ctx context.Context, client Client) (Client, error)

	// UpdateTags updates the client tags.
	UpdateTags(ctx context.Context, client Client) (Client, error)

	// UpdateIdentity updates identity for client with given id.
	UpdateIdentity(ctx context.Context, client Client) (Client, error)

	// UpdateSecret updates secret for client with given identity.
	UpdateSecret(ctx context.Context, client Client) (Client, error)

	// UpdateOwner updates owner for client with given id.
	UpdateOwner(ctx context.Context, client Client) (Client, error)

	// ChangeStatus changes client status to enabled or disabled
	ChangeStatus(ctx context.Context, id string, status Status) (Client, error)
}

// ClientService specifies an API that must be fullfiled by the domain service
// implementation, and all of its decorators (e.g. logging & metrics).
type ClientService interface {
	// RegisterClient creates new client. In case of the failed registration, a
	// non-nil error value is returned.
	RegisterClient(ctx context.Context, token string, client Client) (Client, error)

	// LoginClient authenticates the client given its credentials. Successful
	// authentication generates new access token. Failed invocations are
	// identified by the non-nil error values in the response.

	// ViewClient retrieves client info for a given client ID and an authorized token.
	ViewClient(ctx context.Context, token, id string) (Client, error)

	// ViewProfile retrieves client info for a given token.
	ViewProfile(ctx context.Context, token string) (Client, error)

	// ListClients retrieves clients list for a valid auth token.
	ListClients(ctx context.Context, token string, pm Page) (ClientsPage, error)

	// ListMembers retrieves everything that is assigned to a group identified by groupID.
	ListMembers(ctx context.Context, token, groupID string, pm Page) (MembersPage, error)

	// UpdateClient updates the client's name and metadata.
	UpdateClient(ctx context.Context, token string, client Client) (Client, error)

	// UpdateClientTags updates the client's tags.
	UpdateClientTags(ctx context.Context, token string, client Client) (Client, error)

	// UpdateClientIdentity updates the client's identity
	UpdateClientIdentity(ctx context.Context, token, id, identity string) (Client, error)

	// GenerateResetToken email where mail will be sent.
	// host is used for generating reset link.
	GenerateResetToken(ctx context.Context, email, host string) error

	// UpdateClientSecret updates the client's secret
	UpdateClientSecret(ctx context.Context, token, oldSecret, newSecret string) (Client, error)

	// ResetSecret change users secret in reset flow.
	// token can be authentication token or secret reset token.
	ResetSecret(ctx context.Context, resetToken, secret string) error

	// SendPasswordReset sends reset password link to email.
	SendPasswordReset(ctx context.Context, host, email, token string) error

	// UpdateClientOwner updates the client's owner.
	UpdateClientOwner(ctx context.Context, token string, client Client) (Client, error)

	// EnableClient logically enableds the client identified with the provided ID
	EnableClient(ctx context.Context, token, id string) (Client, error)

	// DisableClient logically disables the client identified with the provided ID
	DisableClient(ctx context.Context, token, id string) (Client, error)

	// Identify returns the client email and id from the given token
	Identify(ctx context.Context, tkn string) (UserIdentity, error)
}

// Custom Marshaller for Client
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

// Custom Unmarshaller for Client
func (s *Status) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), "\"")
	val, err := ToStatus(str)
	*s = val
	return err
}

// Validate returns an error if client representation is invalid.
func (u Client) Validate() error {
	if !isEmail(u.Credentials.Identity) {
		return errors.ErrMalformedEntity
	}
	return nil
}

func isEmail(email string) bool {
	if email == "" {
		return false
	}

	es := strings.Split(email, atSeparator)
	if len(es) != 2 {
		return false
	}
	local, host := es[0], es[1]

	if local == "" || len(local) > maxLocalLen {
		return false
	}

	hs := strings.Split(host, dotSeparator)
	if len(hs) < 2 {
		return false
	}
	domain, ext := hs[0], hs[1]

	// Check subdomain and validate
	if len(hs) > 2 {
		if domain == "" {
			return false
		}

		for i := 1; i < len(hs)-1; i++ {
			sub := hs[i]
			if sub == "" {
				return false
			}
			domain = fmt.Sprintf("%s.%s", domain, sub)
		}

		ext = hs[len(hs)-1]
	}

	if domain == "" || len(domain) > maxDomainLen {
		return false
	}
	if ext == "" || len(ext) > maxTLDLen {
		return false
	}

	punyLocal, err := idna.ToASCII(local)
	if err != nil {
		return false
	}
	punyHost, err := idna.ToASCII(host)
	if err != nil {
		return false
	}

	if userDotRegexp.MatchString(punyLocal) || !userRegexp.MatchString(punyLocal) || !hostRegexp.MatchString(punyHost) {
		return false
	}

	return true
}
