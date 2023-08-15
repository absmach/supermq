package modbus

import (
	"encoding/hex"
	"testing"

	"github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {
	modbusService, err := NewTCPClient(TCPHandlerOptions{
		Address: Address,
	})
	if err != nil {
		t.Fatalf("Failed to create ModbusService: %v", err)
	}
	defer modbusService.Close()

	tests := []struct {
		name          string
		readOpts      RWOptions
		result        string
		err           error
		exceptionCode byte
		dataPointOpt  dataPoint
	}{
		{
			name: "Test Read Holding Register",
			readOpts: RWOptions{
				Address:  100,
				Quantity: 1,
			},
			result:       "ff00",
			err:          nil,
			dataPointOpt: HoldingRegister,
		},
		{
			name: "Test Read Holding Register with error",
			readOpts: RWOptions{
				Address:  201,
				Quantity: 1,
			},
			result:        "",
			exceptionCode: 0x1, // illegal action.
			dataPointOpt:  HoldingRegister,
		},
		{
			name: "Test invalid input",
			readOpts: RWOptions{
				Address:  101,
				Quantity: 1,
			},
			result:       "",
			err:          errUnsupportedRead,
			dataPointOpt: Register,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			readData, err := modbusService.Read(test.readOpts.Address, test.readOpts.Quantity, test.dataPointOpt)
			switch er := err.(type) {
			case *modbus.ModbusError:
				assert.Equal(t, test.exceptionCode, er.ExceptionCode)
			default:
				assert.Equal(t, test.err, err)
			}
			assert.Equal(t, test.result, hex.EncodeToString(readData))
		})
	}
}

func TestWrite(t *testing.T) {
	modbusService, err := NewTCPClient(TCPHandlerOptions{
		Address: Address,
	})
	if err != nil {
		t.Fatalf("Failed to create ModbusService: %v", err)
	}
	defer modbusService.Close()

	tests := []struct {
		name          string
		writeOpts     RWOptions
		result        string
		err           error
		exceptionCode byte
		dataPointOpt  dataPoint
	}{
		{
			name: "Test Write Single Register",
			writeOpts: RWOptions{
				Address:  100,
				Quantity: 1,
				Value: ValueWrapper{
					Data: uint16(1),
				},
			},
			dataPointOpt: Register,
			err:          nil,
			result:       "0001",
		},
		{
			name: "Test Write Single Register with invalid input",
			writeOpts: RWOptions{
				Address:  100,
				Quantity: 1,
				Value: ValueWrapper{
					Data: 1,
				},
			},
			dataPointOpt: Register,
			err:          errInvalidInput,
			result:       "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := modbusService.Write(test.writeOpts.Address, test.writeOpts.Quantity, test.writeOpts.Value.Data, test.dataPointOpt)
			switch er := err.(type) {
			case *modbus.ModbusError:
				assert.Equal(t, test.exceptionCode, er.ExceptionCode)
			default:
				assert.Equal(t, test.err, err)
			}
			assert.Equal(t, test.result, hex.EncodeToString(res))
		})
	}
}
