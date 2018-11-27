package redis

type createThingEvent struct {
	id       string
	owner    string
	kind     string
	name     string
	metadata string
}

type updateThingEvent struct {
	id       string
	kind     string
	name     string
	metadata string
}

type removeThingEvent struct {
	id string
}

type createChannelEvent struct {
	id    string
	owner string
	name  string
}

type updateChannelEvent struct {
	id   string
	name string
}

type removeChannelEvent struct {
	id string
}

type connectThingEvent struct {
	thingID string
	chanID  string
}

type disconnectThingEvent struct {
	thingID string
	chanID  string
}
