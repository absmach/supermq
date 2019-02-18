package redis

type removeEvent struct {
	id string
}

type updateChannelEvent struct {
	id       string
	name     string
	metadata string
}
