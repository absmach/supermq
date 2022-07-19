// Copyright (c) Mainflux
// SPDX-Licence-Identifier: Apache-2.0

package api

import "github.com/gorilla/websocket"

type publishReq struct {
	thingKey string // pubID = thingKey (delete this comment later)
	chanID   string
	subtopic string
	conn     *websocket.Conn
}
