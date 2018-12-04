package bootstrap

// IdentityProvider specifies an API for identity management via security tokens.
type IdentityProvider interface {
	ExtractKey(string) (string, error)
	Identify(string) (string, error)
}
