package modbus

import (
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"time"
)

// function codes
const (
	READ_COIL                       = 0x01
	READ_DISCRETE_INPUT             = 0x02
	READ_MULTIPLE_HOLDING_REGISTERS = 0x03
	READ_INPUT_REGISTERS            = 0x04

	WRITE_SINGLE_COIL                = 0x05
	WRITE_SINGLE_HOLDING_REGISTERS   = 0x06
	WRITE_MULTIPLE_COILS             = 0x15
	WRITE_MULTIPLE_HOLDING_REGISTERS = 0x16
)

// exception codes
const (
	ERR_ILLEGAL_FUNCTION                        = 0x01
	ERR_ILLEGAL_DATA_ADDRESS                    = 0x02
	ERR_ILLEGAL_DATA_VALUE                      = 0x03
	ERR_SLAVE_DEVICE_FAILURE                    = 0x04
	ERR_ACKNOWLEDGE                             = 0x05
	ERR_SLAVE_DEVICE_BUSY                       = 0x06
	ERR_NEGATIVE_ACKNOWLEDGE                    = 0x07
	ERR_MEMORY_PARITY_ERROR                     = 0x08
	ERR_GATEWAY_PATH_UNAVAILABLE                = 0x10
	ERR_GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND = 0x11
)

var modbus_exception = map[int]string{
	ERR_ILLEGAL_FUNCTION:                        "ILLEGAL_FUNCTION",
	ERR_ILLEGAL_DATA_ADDRESS:                    "ILLEGAL_DATA_ADDRESS",
	ERR_ILLEGAL_DATA_VALUE:                      "ILLEGAL_DATA_VALUE",
	ERR_SLAVE_DEVICE_FAILURE:                    "SLAVE_DEVICE_FAILURE",
	ERR_ACKNOWLEDGE:                             "ACKNOWLEDGE",
	ERR_SLAVE_DEVICE_BUSY:                       "SLAVE_DEVICE_BUSY",
	ERR_NEGATIVE_ACKNOWLEDGE:                    "NEGATIVE_ACKNOWLEDGE",
	ERR_MEMORY_PARITY_ERROR:                     "MEMORY_PARITY_ERROR",
	ERR_GATEWAY_PATH_UNAVAILABLE:                "GATEWAY_PATH_UNAVAILABLE",
	ERR_GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND: "GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND",
}

func ConnectTCP(ctx context.Context, addr string, slave_id uint8, start_register int, qty uint16) error {
	_, err := netip.ParseAddrPort(addr)
	if err != nil {
		return fmt.Errorf("failed to parse address: %v", err)
	}

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	var tx_id uint16

	timer := time.NewTimer(time.Second)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			fmt.Print("\033c")
			fmt.Printf("Connection\t: %s\tStart Register\t: %d\n", addr, start_register)
			fmt.Printf("Device ID\t: %d\t\t\tCount\t\t: %d\n", slave_id, qty)
			fmt.Printf("Function Code\t: 0x03 (READ_MULTIPLE_HOLDING_REGISTERS)\n")
			result, err := ReadHoldingRegistersTCP(conn, tx_id, slave_id, start_register, qty)
			if err != nil {
				fmt.Printf("[ERROR] %v\n", err)
			} else {
				for idx := range qty {
					register := start_register + int(idx)
					var bytes strings.Builder
					for idx, val := range hex.EncodeToString(result[register]) {
						if idx != 0 && idx%2 == 0 {
							bytes.WriteRune(' ')
						}
						bytes.WriteRune(val)
					}

					fmt.Printf("%d: %5d [%s]\n", register, binary.BigEndian.Uint16(result[register]), bytes.String())
				}
			}
			tx_id++
			timer.Reset(time.Second)
		}
	}
}

func ReadHoldingRegistersTCP(conn net.Conn, tx_id uint16, slave_id uint8, start_addr int, qty uint16) (map[int][]byte, error) {
	request := FrameFC03TCP(tx_id, slave_id, ConvertRegisterAddr(start_addr), qty)
	if _, err := conn.Write(request); err != nil {
		return nil, err
	}

	// 7 tcp header, 2 modbus header, 2 bytes per register
	var response = make([]byte, 7+2+(2*qty))
	if _, err := conn.Read(response); err != nil {
		return nil, err
	}
	if tx_id != binary.BigEndian.Uint16(response[0:2]) {
		return nil, fmt.Errorf("difference in tx_id %d, got %d", tx_id, binary.BigEndian.Uint16(response[0:2]))
	}
	if slave_id != response[6] {
		return nil, fmt.Errorf("difference in slave_id %d, got %d", slave_id, response[6])
	}
	if READ_MULTIPLE_HOLDING_REGISTERS != response[7] {
		return nil, fmt.Errorf("exception: %s", modbus_exception[int(response[8])])
	}

	var result = make(map[int][]byte, qty)
	for idx := range qty {
		result[start_addr+int(idx)] = response[(9 + (idx * 2)):][:2]
	}

	return result, nil
}

// FrameFC03TCP returns modbus tcp frame for READ_MULTIPLE_HOLDING_REGISTERS
func FrameFC03TCP(tx_id uint16, slave_id uint8, start_addr uint16, qty uint16) []byte {
	req := make([]byte, 12)
	binary.BigEndian.PutUint16(req[0:], tx_id)  // transaction_id
	binary.BigEndian.PutUint16(req[2:], 0x0000) // protocol_id
	binary.BigEndian.PutUint16(req[4:], 6)      // length
	req[6] = slave_id
	req[7] = READ_MULTIPLE_HOLDING_REGISTERS
	binary.BigEndian.PutUint16(req[8:], start_addr)
	binary.BigEndian.PutUint16(req[10:], qty)
	return req
}

// FrameFC03RTU returns modbus rtu frame for READ_MULTIPLE_HOLDING_REGISTERS
func FrameFC03RTU(slave_id uint8, start_addr, qty uint16) []byte {
	req := make([]byte, 8)
	req[0] = slave_id
	req[1] = READ_MULTIPLE_HOLDING_REGISTERS
	binary.BigEndian.PutUint16(req[2:4], start_addr)
	binary.BigEndian.PutUint16(req[4:6], qty)
	binary.LittleEndian.PutUint16(req[6:8], calculateCRCFast(req[:6]))
	return req
}

// ConvertRegisterAddr converts documentation notation 4xxxx/4xxxxxx address to 0 indexed protocol address
func ConvertRegisterAddr(addr int) uint16 {
	if addr > 4_00000 && addr <= 4_65536 {
		return uint16(addr - 4_00001)
	} else if addr > 4_0000 {
		return uint16(addr - 4_0001)
	}
	panic("register address should be 400000-465536 or 40000-49999")
}

// Fast MODBUS CRC 16-bit algorithm in Golang.
// Author: Krzysztof Żerebecki
// This algorithm uses the precomputed CRC values for all possible 8-bit input values (0x00 to 0xFF)
// Computational complexity: O(n) ~ n
// little endian
func calculateCRCFast(data []byte) uint16 {
	const initialCRCValue = 0xFFFF

	table := [256]uint16{
		0x0000, 0xC0C1, 0xC181, 0x0140, 0xC301, 0x03C0, 0x0280, 0xC241,
		0xC601, 0x06C0, 0x0780, 0xC741, 0x0500, 0xC5C1, 0xC481, 0x0440,
		0xCC01, 0x0CC0, 0x0D80, 0xCD41, 0x0F00, 0xCFC1, 0xCE81, 0x0E40,
		0x0A00, 0xCAC1, 0xCB81, 0x0B40, 0xC901, 0x09C0, 0x0880, 0xC841,
		0xD801, 0x18C0, 0x1980, 0xD941, 0x1B00, 0xDBC1, 0xDA81, 0x1A40,
		0x1E00, 0xDEC1, 0xDF81, 0x1F40, 0xDD01, 0x1DC0, 0x1C80, 0xDC41,
		0x1400, 0xD4C1, 0xD581, 0x1540, 0xD701, 0x17C0, 0x1680, 0xD641,
		0xD201, 0x12C0, 0x1380, 0xD341, 0x1100, 0xD1C1, 0xD081, 0x1040,
		0xF001, 0x30C0, 0x3180, 0xF141, 0x3300, 0xF3C1, 0xF281, 0x3240,
		0x3600, 0xF6C1, 0xF781, 0x3740, 0xF501, 0x35C0, 0x3480, 0xF441,
		0x3C00, 0xFCC1, 0xFD81, 0x3D40, 0xFF01, 0x3FC0, 0x3E80, 0xFE41,
		0xFA01, 0x3AC0, 0x3B80, 0xFB41, 0x3900, 0xF9C1, 0xF881, 0x3840,
		0x2800, 0xE8C1, 0xE981, 0x2940, 0xEB01, 0x2BC0, 0x2A80, 0xEA41,
		0xEE01, 0x2EC0, 0x2F80, 0xEF41, 0x2D00, 0xEDC1, 0xEC81, 0x2C40,
		0xE401, 0x24C0, 0x2580, 0xE541, 0x2700, 0xE7C1, 0xE681, 0x2640,
		0x2200, 0xE2C1, 0xE381, 0x2340, 0xE101, 0x21C0, 0x2080, 0xE041,
		0xA001, 0x60C0, 0x6180, 0xA141, 0x6300, 0xA3C1, 0xA281, 0x6240,
		0x6600, 0xA6C1, 0xA781, 0x6740, 0xA501, 0x65C0, 0x6480, 0xA441,
		0x6C00, 0xACC1, 0xAD81, 0x6D40, 0xAF01, 0x6FC0, 0x6E80, 0xAE41,
		0xAA01, 0x6AC0, 0x6B80, 0xAB41, 0x6900, 0xA9C1, 0xA881, 0x6840,
		0x7800, 0xB8C1, 0xB981, 0x7940, 0xBB01, 0x7BC0, 0x7A80, 0xBA41,
		0xBE01, 0x7EC0, 0x7F80, 0xBF41, 0x7D00, 0xBDC1, 0xBC81, 0x7C40,
		0xB401, 0x74C0, 0x7580, 0xB541, 0x7700, 0xB7C1, 0xB681, 0x7640,
		0x7200, 0xB2C1, 0xB381, 0x7340, 0xB101, 0x71C0, 0x7080, 0xB041,
		0x5000, 0x90C1, 0x9181, 0x5140, 0x9301, 0x53C0, 0x5280, 0x9241,
		0x9601, 0x56C0, 0x5780, 0x9741, 0x5500, 0x95C1, 0x9481, 0x5440,
		0x9C01, 0x5CC0, 0x5D80, 0x9D41, 0x5F00, 0x9FC1, 0x9E81, 0x5E40,
		0x5A00, 0x9AC1, 0x9B81, 0x5B40, 0x9901, 0x59C0, 0x5880, 0x9841,
		0x8801, 0x48C0, 0x4980, 0x8941, 0x4B00, 0x8BC1, 0x8A81, 0x4A40,
		0x4E00, 0x8EC1, 0x8F81, 0x4F40, 0x8D01, 0x4DC0, 0x4C80, 0x8C41,
		0x4400, 0x84C1, 0x8581, 0x4540, 0x8701, 0x47C0, 0x4680, 0x8641,
		0x8201, 0x42C0, 0x4380, 0x8341, 0x4100, 0x81C1, 0x8081, 0x4040}

	crc := uint16(initialCRCValue)
	var xor uint8

	for _, b := range data {
		xor = b ^ uint8(crc)
		crc >>= 8
		crc ^= table[xor]
	}

	return crc
}
