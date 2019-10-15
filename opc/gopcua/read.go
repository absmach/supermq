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

// OPCReader represents the OPC client.
type OPCReader interface {
	// Read given NodeID.
	Read(string, string) error
}

type reader struct {
	ctx    context.Context
	svc    opc.Service
	logger logger.Logger
}

// NewReader returns new OPC reader instance.
func NewReader(ctx context.Context, svc opc.Service, log logger.Logger) OPCReader {
	return reader{
		ctx:    ctx,
		svc:    svc,
		logger: log,
	}
}

// Read reads a given OPC-UA Server endpoint.
func (r reader) Read(endpoint, nodeID string) error {
	c := opcua.NewClient(endpoint, opcua.SecurityMode(ua.MessageSecurityModeNone))
	if err := c.Connect(r.ctx); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	id, err := ua.ParseNodeID(nodeID)
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
		Namespace: defOPCNamespace,
		ID:        defOPCIdentifier,
		Data:      resp.Results[0].Value.Float(),
	}
	r.svc.Publish(r.ctx, "", msg)

	return nil
}
