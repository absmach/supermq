package modbus

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/messaging"
)

const (
	// channels.<channel_id>.modbus.<read/write>.<modbus_protocol>.<modbus_data_point>
	readTopic  = "channels.*.modbus.read.*.*"
	writeTopic = "channels.*.modbus.write.*.*"
)

var errUnsupportedModbusProtocol = errors.New("unsupported modbus protocol")

type Service interface {
	// Read subscribes to the Subscriber and
	// reads modbus sensor values while publishing them to publisher.
	Read(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error
	// Write subscribes to the Subscriber and
	// writes to modbus sensor.
	Write(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error
}

type service struct {
	logger mflog.Logger
}

// NewForwarder returns new Forwarder implementation.
func New(logger mflog.Logger) Service {
	return service{
		logger: logger,
	}
}

func (f service) Read(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error {
	return sub.Subscribe(ctx, id, readTopic, handleRead(ctx, pub, f.logger))
}

func handleRead(ctx context.Context, pub messaging.Publisher, logger mflog.Logger) handleFunc {
	return func(msg *messaging.Message) error {
		protocal := strings.Split(msg.Subtopic, ".")[2]
		dp := strings.Split(msg.Subtopic, ".")[3]
		writeOpts, cfg, err := getInput(msg.Payload)
		client, err := clientFromProtocol(protocal, cfg)
		if err != nil {
			return err
		}
		go func() {
			defer client.Close()
			for {
				res, err := client.Read(writeOpts.Address, writeOpts.Quantity, dataPoint(dp))
				if err != nil {
					logger.Error(err.Error())
				}
				if err := pub.Publish(ctx, msg.Channel, &messaging.Message{
					Payload:  res,
					Subtopic: "modbus.res",
				}); err != nil {
					logger.Error(err.Error())
				}
			}
		}()
		return nil
	}
}

func (f service) Write(ctx context.Context, id string, sub messaging.Subscriber, pub messaging.Publisher) error {
	return sub.Subscribe(ctx, id, writeTopic, handleWrite(ctx, pub, f.logger))
}

func handleWrite(ctx context.Context, pub messaging.Publisher, logger mflog.Logger) handleFunc {
	return func(msg *messaging.Message) error {
		protocal := strings.Split(msg.Subtopic, ".")[2]
		dp := strings.Split(msg.Subtopic, ".")[3]
		writeOpts, cfg, err := getInput(msg.Payload)
		client, err := clientFromProtocol(protocal, cfg)
		if err != nil {
			return err
		}
		defer client.Close()
		res, err := client.Write(writeOpts.Address, writeOpts.Quantity, writeOpts.Value.Data, dataPoint(dp))
		if err != nil {
			return err
		}
		if err := pub.Publish(ctx, msg.Channel, &messaging.Message{
			Payload:  res,
			Subtopic: "modbus.res",
		}); err != nil {
			return err
		}

		return nil
	}
}

type handleFunc func(msg *messaging.Message) error

func (h handleFunc) Handle(msg *messaging.Message) error {
	return h(msg)

}

func (h handleFunc) Cancel() error {
	return nil
}

func clientFromProtocol(protocol string, config []byte) (ModbusService, error) {
	switch protocol {
	case "tcp":
		var cfg TCPHandlerOptions
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, err
		}
		return NewTCPClient(cfg)
	case "rtu":
		var cfg RTUHandlerOptions
		if err := json.Unmarshal(config, &cfg); err != nil {
			return nil, err
		}
		return NewRTUClient(cfg)
	default:
		return nil, errUnsupportedModbusProtocol
	}
}

func getInput(data []byte) (RWOptions, []byte, error) {
	var opts RWOptions
	var confs json.RawMessage
	if err := json.Unmarshal(data, &struct {
		Options *RWOptions       `json:"options"`
		Config  *json.RawMessage `json:"config"`
	}{
		Options: &opts,
		Config:  &confs,
	}); err != nil {
		return RWOptions{}, nil, nil
	}
	return opts, confs, nil
}
