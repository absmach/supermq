// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package gopcua

import (
	"context"

	opcuaGopcua "github.com/gopcua/opcua"
	uaGopcua "github.com/gopcua/opcua/ua"
	"github.com/mainflux/mainflux/errors"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/opcua"
)

var _ opcua.Reader = (*reader)(nil)

type reader struct {
	ctx    context.Context
	svc    opcua.Service
	logger logger.Logger
}

// NewReader returns new OPC-UA reader instance.
func NewReader(ctx context.Context, svc opcua.Service, log logger.Logger) opcua.Reader {
	return reader{
		ctx:    ctx,
		svc:    svc,
		logger: log,
	}
}

// Read reads a given OPC-UA Server endpoint.
func (r reader) Read(cfg opcua.Config) error {
	c := opcuaGopcua.NewClient(cfg.ServerURI, opcuaGopcua.SecurityMode(uaGopcua.MessageSecurityModeNone))
	if err := c.Connect(r.ctx); err != nil {
		return errors.Wrap(errFailedConn, err)
	}
	defer c.Close()

	id, err := uaGopcua.ParseNodeID(cfg.NodeID)
	if err != nil {
		return errors.Wrap(errFailedParseNodeID, err)
	}

	req := &uaGopcua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*uaGopcua.ReadValueID{
			&uaGopcua.ReadValueID{NodeID: id},
		},
		TimestampsToReturn: uaGopcua.TimestampsToReturnBoth,
	}

	resp, err := c.Read(req)
	if err != nil {
		return errors.Wrap(errFailedRead, err)
	}
	if resp.Results[0].Status != uaGopcua.StatusOK {
		return errResponseStatus
	}

	// Publish on Mainflux NATS broker
	msg := opcua.Message{
		ServerURI: cfg.ServerURI,
		NodeID:    cfg.NodeID,
		Data:      resp.Results[0].Value.Float(),
	}
	r.svc.Publish(r.ctx, "", msg)

	return nil
}
