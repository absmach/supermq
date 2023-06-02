package grpc

type identityRes struct {
	id string
}

type authorizeRes struct {
	thingID    string
	authorized bool
}
