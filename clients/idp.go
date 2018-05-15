package clients

// IdentityProvider specifies an API for identity management via security
// tokens.
type IdentityProvider interface {
	// Key generates the non-expiring access token.
	Key() string

	// Identity extracts the entity identifier given its secret key.
	Identity(string) (string, error)
}
