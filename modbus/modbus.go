package modbus

import (
	"log"
	"reflect"
	"time"

	"github.com/goburrow/modbus"
	"github.com/goburrow/serial"
)

// TCPHandlerOptions defines optional handler values.
type TCPHandlerOptions struct {
	IdleTimeout time.Duration
	Logger      *log.Logger
	SlaveId     byte
	Timeout     time.Duration
}

// NewRTUClient initializes a new modbus.Client on TCP protocol from the address
// and handler options provided.
func NewTCPClient(address string, config TCPHandlerOptions) (modbus.Client, error) {
	handler := modbus.NewTCPClientHandler(address)
	if err := handler.Connect(); err != nil {
		return nil, err
	}
	if !isZeroValue(config.IdleTimeout) {
		handler.IdleTimeout = config.IdleTimeout
	}
	if !isZeroValue(config.Logger) {
		handler.Logger = config.Logger
	}
	if !isZeroValue(config.SlaveId) {
		handler.SlaveId = config.SlaveId
	}
	if !isZeroValue(config.Timeout) {
		handler.Timeout = config.Timeout
	}
	return modbus.NewClient(handler), nil
}

// RTUHandlerOptions defines optional handler values.
type RTUHandlerOptions struct {
	BaudRate    int
	Config      serial.Config
	DataBits    int
	IdleTimeout time.Duration
	Logger      *log.Logger
	Parity      string
	RS485       serial.RS485Config
	SlaveId     byte
	StopBits    int
	Timeout     time.Duration
}

// NewRTUClient initializes a new modbus.Client on RTU/ASCII protocol from the address
// and handler options provided.
func NewRTUClient(address string, config RTUHandlerOptions) (modbus.Client, error) {
	handler := modbus.NewRTUClientHandler(address)
	if err := handler.Connect(); err != nil {
		return nil, err
	}
	if !isZeroValue(config.BaudRate) {
		handler.BaudRate = config.BaudRate
	}
	if !isZeroValue(config.Config) {
		handler.Config = config.Config
	}
	if !isZeroValue(config.DataBits) {
		handler.DataBits = config.DataBits
	}
	if !isZeroValue(config.IdleTimeout) {
		handler.IdleTimeout = config.IdleTimeout
	}
	if !isZeroValue(config.Logger) {
		handler.Logger = config.Logger
	}
	if !isZeroValue(config.Parity) {
		handler.Parity = config.Parity
	}
	if !isZeroValue(config.RS485) {
		handler.RS485 = config.RS485
	}
	if !isZeroValue(config.SlaveId) {
		handler.SlaveId = config.SlaveId
	}
	if !isZeroValue(config.StopBits) {
		handler.StopBits = config.StopBits
	}
	if !isZeroValue(config.Timeout) {
		handler.Timeout = config.Timeout
	}
	return modbus.NewClient(handler), nil
}

func isZeroValue(val interface{}) bool {
	v := reflect.ValueOf(val)
	switch v.Kind() {
	case reflect.Func, reflect.Map, reflect.Slice:
		return v.IsNil()
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	default:
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}
}
