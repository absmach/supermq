// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package gopcua

// LoraSubscribe subscribe to opc server messages
import (
	"context"
	"fmt"
	"log"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/opc"
)

// Reader represents the OPC client.
type Reader interface {
	// Read given OPC-UA Server NodeID (Namespace + ID).
	Read(opc.Config) error
}

type reader struct {
	ctx    context.Context
	svc    opc.Service
	logger logger.Logger
}

// NewReader returns new OPC reader instance.
func NewReader(ctx context.Context, svc opc.Service, log logger.Logger) Reader {
	return reader{
		ctx:    ctx,
		svc:    svc,
		logger: log,
	}
}

// Read reads a given OPC-UA Server endpoint.
func (r reader) Read(oc opc.Config) error {
	c := opcua.NewClient(oc.ServerURI, opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err := c.Connect(r.ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	nid := fmt.Sprintf("ns=%s;i=%s", oc.NodeNamespace, oc.NodeIdintifier)
	id, err := ua.ParseNodeID(nid)
	if err != nil {
		r.logger.Error(fmt.Sprintf("invalid node id: %v", err))
	}

	req := &ua.ReadRequest{
		MaxAge: 2000,
		NodesToRead: []*ua.ReadValueID{
			&ua.ReadValueID{NodeID: id},
		},
		TimestampsToReturn: ua.TimestampsToReturnBoth,
	}

	resp, err := c.Read(req)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Read failed: %s", err))
	}
	if resp.Results[0].Status != ua.StatusOK {
		r.logger.Error(fmt.Sprintf("Status not OK: %v", resp.Results[0].Status))
	}

	// Publish on Mainflux NATS broker
	msg := opc.Message{
		Namespace: oc.NodeNamespace,
		ID:        oc.NodeIdintifier,
		Data:      resp.Results[0].Value.Float(),
	}
	r.svc.Publish(r.ctx, "", msg)

	return nil
}
