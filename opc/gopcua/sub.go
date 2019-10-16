// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package gopcua

// LoraSubscribe subscribe to opc server messages
import (
	"context"
	"fmt"
	"time"

	"github.com/gopcua/opcua"
	"github.com/gopcua/opcua/ua"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/opc"
)

const (
	defOPCNamespace  = "0"
	defOPCIdentifier = "2256"
)

// Subscriber represents the OPC-UA Server client.
type Subscriber interface {
	// Subscribes to given NodeID and receives events.
	Subscribe(string, string) error
}

type client struct {
	ctx    context.Context
	svc    opc.Service
	logger logger.Logger
}

// NewClient returns new OPC client instance.
func NewClient(ctx context.Context, svc opc.Service, log logger.Logger) Subscriber {
	return client{
		ctx:    ctx,
		svc:    svc,
		logger: log,
	}
}

// Subscribe subscribes to the OPC-UA Server.
func (b client) Subscribe(uri, nid string) error {
	var (
		policy   = "" // Security policy: None, Basic128Rsa15, Basic256, Basic256Sha256. Default: auto
		mode     = "" // Modes: None, Sign, SignAndEncrypt. Default: auto
		certFile = ""
		keyFile  = ""
	)

	endpoints, err := opcua.GetEndpoints(uri)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to fetch OPC server endpoints: %s", err.Error()))
	}

	ep := opcua.SelectEndpoint(endpoints, policy, ua.MessageSecurityModeFromString(mode))
	if ep == nil {
		b.logger.Error("Failed to find suitable endpoint")
	}

	opts := []opcua.Option{
		opcua.SecurityPolicy(policy),
		opcua.SecurityModeString(mode),
		opcua.CertificateFile(certFile),
		opcua.PrivateKeyFile(keyFile),
		opcua.AuthAnonymous(),
		opcua.SecurityFromEndpoint(ep, ua.UserTokenTypeAnonymous),
	}

	c := opcua.NewClient(ep.EndpointURL, opts...)
	if errC := c.Connect(b.ctx); err != nil {
		b.logger.Error(errC.Error())
	}
	defer c.Close()

	sub, err := c.Subscribe(&opcua.SubscriptionParameters{
		Interval: 2000 * time.Millisecond,
	})
	if err != nil {
		b.logger.Error(err.Error())
	}
	defer sub.Cancel()
	b.logger.Info(fmt.Sprintf("OPC-UA server URI: %s", ep.SecurityPolicyURI))
	b.logger.Info(fmt.Sprintf("Created subscription with id %v", sub.SubscriptionID))

	nodeID, err := ua.ParseNodeID(nid)
	if err != nil {
		b.logger.Error(err.Error())
	}

	// arbitrary client handle for the monitoring item
	handle := uint32(42)
	miCreateRequest := opcua.NewMonitoredItemCreateRequestWithDefaults(nodeID, ua.AttributeIDValue, handle)
	res, err := sub.Monitor(ua.TimestampsToReturnBoth, miCreateRequest)
	if err != nil || res.Results[0].StatusCode != ua.StatusOK {
		b.logger.Error(err.Error())
	}

	go sub.Run(b.ctx)

	for {
		select {
		case <-b.ctx.Done():
			return nil
		case res := <-sub.Notifs:
			if res.Error != nil {
				b.logger.Error(res.Error.Error())
				continue
			}

			switch x := res.Value.(type) {
			case *ua.DataChangeNotification:
				for _, item := range x.MonitoredItems {
					// Publish on Mainflux NATS broker
					msg := opc.Message{
						Namespace: defOPCNamespace,
						ID:        defOPCIdentifier,
						Data:      item.Value.Value.Float(),
					}
					b.svc.Publish(b.ctx, "", msg)
				}

			default:
				b.logger.Info(fmt.Sprintf("what's this publish result? %T", res.Value))
			}
		}
	}
}
