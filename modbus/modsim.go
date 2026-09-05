package modbus

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"sync"
)

// https://www.modbus.org/file/secure/messagingimplementationguide.pdf

type Device struct {
	mutex sync.Mutex

	SlaveID          int
	HoldingRegisters []uint16
	CoilRegisters    []bool
}

func NewModbusDevice(slave_id int, holding_registers []uint16, coil_registers []bool) Device {
	return Device{
		mutex:            sync.Mutex{},
		SlaveID:          slave_id,
		HoldingRegisters: holding_registers,
		CoilRegisters:    coil_registers,
	}
}

func ServeTCP(ctx context.Context, addr string, devices []Device) error {
	mapped := make(map[int]*Device, len(devices))
	for idx := range devices {
		mapped[devices[idx].SlaveID] = &devices[idx]
	}

	_, err := netip.ParseAddrPort(addr)
	if err != nil {
		return fmt.Errorf("failed to parse address: %v", err)
	}
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return fmt.Errorf("failed to start tcp server: %v", err)
	}

	go func() {
		<-ctx.Done()
		listener.Close()
	}()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("[ERROR] failed to accept conn: %v\n", err)
			continue
		}

		tcp_conn, ok := conn.(*net.TCPConn)
		if !ok {
			fmt.Printf("[ERROR] invalid connection type\n")
			continue
		}

		go handleConnection(tcp_conn, mapped)
	}
}

func handleConnection(conn *net.TCPConn, devices map[int]*Device) {
	defer func() {
		conn.Close()
		fmt.Println("[INFO] connection closed")
	}()
	_ = conn.SetKeepAlive(true)

	var read = make([]byte, 255)
	var response []byte
	for {
		n, err := conn.Read(read)
		if err != nil {
			return
		}
		request := read[:n]

		if len(request) < 8 {
			response = append(request[0:8], ERR_SLAVE_DEVICE_FAILURE)
			response[7] += 0x80
		} else {
			function_code := request[7:8][0]
			switch function_code {
			case READ_COIL:
				response = ResponseFC01TCP(request, devices)
			case READ_HOLDING_REGISTERS:
				response = ResponseFC03TCP(request, devices)
			default:
				response = append(request[0:8], ERR_ILLEGAL_FUNCTION)
				response[7] += 0x80
			}
		}

		_, err = conn.Write(response)
		if err != nil {
			fmt.Printf("[ERROR] failed to write response: %v\n", err)
			return
		}
	}
}

func ResponseFC01TCP(request []byte, devices map[int]*Device) []byte {
	panic("unimplemented")
}
func ResponseFC03TCP(request []byte, devices map[int]*Device) []byte {
	tx_id := binary.BigEndian.Uint16(request[0:2])
	protocol_id := binary.BigEndian.Uint16(request[2:4])
	// length := binary.BigEndian.Uint16(request[4:6])
	slave_id := request[6:7][0]
	function_code := request[7:8][0]

	if len(request) < 12 {
		response := append(request[0:8], ERR_ILLEGAL_DATA_ADDRESS)
		response[7] += 0x80
		return response
	}

	device, ok := devices[int(slave_id)]
	if !ok {
		response := append(request[0:8], ERR_SLAVE_DEVICE_FAILURE)
		response[7] += 0x80
		return response
	}

	start_addr := binary.BigEndian.Uint16(request[8:10])
	qty := binary.BigEndian.Uint16(request[10:12])
	max_addr := start_addr + qty - 1 // we -1 because register is 1-indexing but go is 0-indexing
	if len(device.HoldingRegisters) <= int(max_addr) {
		response := append(request[0:8], ERR_ILLEGAL_DATA_VALUE)
		response[7] += 0x80
		return response
	}

	var response = make([]byte, 9+(2*qty))
	binary.BigEndian.PutUint16(response[0:2], tx_id)
	binary.BigEndian.PutUint16(response[2:4], protocol_id)
	binary.BigEndian.PutUint16(response[4:6], 2+(qty*2))
	response[6] = slave_id
	response[7] = function_code
	response[8] = byte(qty) * 2

	var idx uint16
	for addr := start_addr; addr <= max_addr; addr++ {
		idx = (addr - start_addr) * 2
		binary.BigEndian.PutUint16(response[(9+idx):], device.HoldingRegisters[addr])
	}

	return response
}
