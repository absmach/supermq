package grpc

type identityRes struct {
	id  string
	err error
}

func (res identityRes) failed() error {
	return res.err
}
