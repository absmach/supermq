package modbus

import (
	"errors"
	"log"
	"reflect"
	"time"

	"github.com/goburrow/modbus"
	"github.com/goburrow/serial"
)

type IO string

const (
	Coil            IO = "coil"
	HoldingRegister IO = "h_register"
	InputRegister   IO = "i_register"
	Register        IO = "register"
	Discrete        IO = "discrete"
	FIFO            IO = "fifo"
)

var errInvalidInput = errors.New("invalid input type")

type ModbusService interface {
	// Read gets data from modbus.
	Read(address, quantity uint16, iotype IO) ([]byte, error)
	// Write writes a value/s on Modbus.
	Write(address, quantity uint16, value interface{}, iotype IO) ([]byte, error)
}

var _ ModbusService = (*modbusService)(nil)

// adapterService provides methods for reading and writing data on Modbus.
type modbusService struct {
	Client modbus.Client
}

// TCPHandlerOptions defines optional handler values.
type TCPHandlerOptions struct {
	IdleTimeout time.Duration
	Logger      *log.Logger
	SlaveId     byte
	Timeout     time.Duration
}

// NewRTUClient initializes a new modbus.Client on TCP protocol from the address
// and handler options provided.
func NewTCPClient(address string, config TCPHandlerOptions) (ModbusService, error) {
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

	return &modbusService{
		Client: modbus.NewClient(handler),
	}, nil
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
func NewRTUClient(address string, config RTUHandlerOptions) (ModbusService, error) {
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
	return &modbusService{
		Client: modbus.NewClient(handler),
	}, nil
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

// Write writes a value/s on Modbus.
func (s *modbusService) Write(address, quantity uint16, value interface{}, iotype IO) ([]byte, error) {
	switch iotype {
	case Coil:
		switch val := value.(type) {
		case uint16:
			return s.Client.WriteSingleCoil(address, val)
		case []byte:
			return s.Client.WriteMultipleCoils(address, quantity, val)
		default:
			return nil, errInvalidInput
		}
	case Register:
		switch val := value.(type) {
		case uint16:
			return s.Client.WriteSingleRegister(address, val)
		case []byte:
			return s.Client.WriteMultipleRegisters(address, quantity, val)
		default:
			return nil, errInvalidInput
		}
	default:
		return nil, errInvalidInput
	}
}

// Read gets data from modbus.
func (s *modbusService) Read(address uint16, quantity uint16, iotype IO) ([]byte, error) {
	switch iotype {
	case Coil:
		return s.Client.ReadCoils(address, quantity)
	case Discrete:
		return s.Client.ReadDiscreteInputs(address, quantity)
	case FIFO:
		return s.Client.ReadFIFOQueue(address)
	case HoldingRegister:
		return s.Client.ReadHoldingRegisters(address, quantity)
	case InputRegister:
		return s.Client.ReadInputRegisters(address, quantity)
	default:
		return nil, errInvalidInput
	}
}
