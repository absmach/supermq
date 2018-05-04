package mainflux

// APIRes contains http response specifig methods.
type APIRes interface {
	Code() int
	Headers() map[string]string
	Empty() bool
}
