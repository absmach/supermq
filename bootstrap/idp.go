package bootstrap

// IdentityProvider specifies an API for identity management via security tokens.
type IdentityProvider interface {

	// ExtractKey parses user token and extracts user key from it.
	ExtractKey(string) (string, error)
	// Identify validates user token using Mainflux API.
	Identify(string) (string, error)
}
