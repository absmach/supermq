// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

type deleteUserRes struct {
	deleted bool
}

type retrieveIDByRouteRes struct {
	id string
}

type retrieveStatusRes struct {
	status uint8
}
