// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

type authorizeRes struct {
	id         string
	authorized bool
}

type getEntitiesBasicRes struct {
	total  uint64
	limit  uint64
	offset uint64
	things []thingBasic
}

type thingBasic struct {
	id     string
	domain string
	status uint8
}

type connectionsReq struct {
	connections []connection
}

type connection struct {
	thingID   string
	channelID string
	domainID  string
}
type connectionsRes struct {
	ok bool
}
