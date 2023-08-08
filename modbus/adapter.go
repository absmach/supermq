package modbus

import (
	"errors"

	"github.com/goburrow/modbus"
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

type Service interface {
	// Read gets data from modbus.
	Read(address, quantity uint16, iotype IO) ([]byte, error)
	// Write writes a value/s on Modbus.
	Write(address, quantity uint16, value interface{}, iotype IO) ([]byte, error)
}

var _ Service = (*adapterService)(nil)

// adapterService provides methods for reading and writing data on Modbus.
type adapterService struct {
	Client modbus.Client
}

// NewModbusService creates a new instance of ModbusService.
func NewModbusService(client modbus.Client) Service {
	return &adapterService{Client: client}
}

// Write writes a value/s on Modbus.
func (s *adapterService) Write(address, quantity uint16, value interface{}, iotype IO) ([]byte, error) {
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
func (s *adapterService) Read(address uint16, quantity uint16, iotype IO) ([]byte, error) {
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
