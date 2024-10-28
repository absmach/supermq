// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

type removeThingConnectionsReq struct {
	thingID string
}

type unsetParentGroupFromChannelsReq struct {
	parentGroupID string
}
