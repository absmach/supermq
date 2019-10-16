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

// Subscriber represents the OPC-UA Server client.
type Subscriber interface {
	// Subscribes to given NodeID and receives events.
	Subscribe(opc.Config) error
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
func (b client) Subscribe(oc opc.Config) error {
	endpoints, err := opcua.GetEndpoints(oc.ServerURI)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to fetch OPC server endpoints: %s", err.Error()))
	}

	ep := opcua.SelectEndpoint(endpoints, oc.Policy, ua.MessageSecurityModeFromString(oc.Mode))
	if ep == nil {
		b.logger.Error("Failed to find suitable endpoint")
	}

	opts := []opcua.Option{
		opcua.SecurityPolicy(oc.Policy),
		opcua.SecurityModeString(oc.Mode),
		opcua.CertificateFile(oc.CertFile),
		opcua.PrivateKeyFile(oc.KeyFile),
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

	if err := b.runHandler(sub, oc); err != nil {
		return err
	}

	return nil
}

func (b client) runHandler(sub *opcua.Subscription, cfg opc.Config) error {
	nid := fmt.Sprintf("ns=%s;i=%s", cfg.NodeNamespace, cfg.NodeIdintifier)
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
						Namespace: cfg.NodeNamespace,
						ID:        cfg.NodeIdintifier,
						Data:      item.Value.Value.Float(),
					}
					b.svc.Publish(b.ctx, "", msg)
				}

			default:
				b.logger.Info(fmt.Sprintf("what's this publish result? %T", res.Value))
			}
		}
	}

	return nil
}
