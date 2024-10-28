// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

type authorizeReq struct {
	ThingID    string
	ThingKey   string
	ChannelID  string
	Permission string
}

type retrieveEntitiesReq struct {
	Ids []string
}

type retrieveEntityReq struct {
	Id string
}

type removeChannelConnectionsReq struct {
	channelID string
}

type unsetParentGroupFormThingsReq struct {
	parentGroupID string
}
