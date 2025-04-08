// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package grpc

import (
	"github.com/absmach/supermq/pkg/connections"
	"github.com/absmach/supermq/pkg/errors"
	"github.com/absmach/supermq/pkg/policies"
)

var errDomainRoute = errors.New("domain route required for users")

type authorizeReq struct {
	domainRoute  string
	channelRoute string
	clientID     string
	clientType   string
	connType     connections.ConnType
}

func (req authorizeReq) validate() error {
	if req.clientType == policies.UserType && req.domainRoute == "" {
		return errDomainRoute
	}
	return nil
}

type removeClientConnectionsReq struct {
	clientID string
}

type unsetParentGroupFromChannelsReq struct {
	parentGroupID string
}

type retrieveEntityReq struct {
	Id string
}
