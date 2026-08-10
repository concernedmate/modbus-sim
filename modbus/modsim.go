package modbus

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/netip"
	"sync"
)

// https://www.modbus.org/file/secure/messagingimplementationguide.pdf

type Device struct {
	mutex             sync.Mutex
	holding_registers []uint16
	coil_registers    []bool
}

func ServeTCP(addr string) error {
	devices := map[int]*Device{}
	devices[1] = &Device{
		mutex:             sync.Mutex{},
		holding_registers: []uint16{128, 255, 65535},
		coil_registers:    []bool{},
	}

	_, err := netip.ParseAddrPort(addr)
	if err != nil {
		return fmt.Errorf("failed to parse address: %v", err)
	}
	listener, err := net.Listen("tcp4", addr)
	if err != nil {
		return fmt.Errorf("failed to start tcp server: %v", err)
	}
	defer listener.Close()

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

		go handleConnection(tcp_conn, devices)
	}
}

func handleConnection(conn *net.TCPConn, devices map[int]*Device) {
	defer func() {
		conn.Close()
		fmt.Println("[INFO] connection closed")
	}()
	_ = conn.SetKeepAlive(true)

	for {
		var request = make([]byte, 255)
		n, err := conn.Read(request)
		if err != nil {
			break
		}
		request = request[:n]

		if len(request) < 8 {
			response := append(request[0:8], ERR_ILLEGAL_FUNCTION)
			response[7] += 0x80
			_, err := conn.Write(response)
			if err != nil {
				fmt.Printf("[ERROR] failed to write response: %v\n", err)
				break
			}
		} else {
			tx_id := binary.BigEndian.Uint16(request[0:2])
			protocol_id := binary.BigEndian.Uint16(request[2:4])
			// length := binary.BigEndian.Uint16(request[4:6])
			slave_id := request[6:7][0]
			function_code := request[7:8][0]

			switch function_code {
			case READ_HOLDING_REGISTERS:
				if len(request) < 12 {
					response := append(request[0:8], ERR_ILLEGAL_DATA_ADDRESS)
					response[7] += 0x80
					_, err := conn.Write(response)
					if err != nil {
						fmt.Printf("[ERROR] failed to write response: %v\n", err)
						break
					}
					continue
				}

				device, ok := devices[int(slave_id)]
				if !ok {
					response := append(request[0:8], ERR_SLAVE_DEVICE_FAILURE)
					response[7] += 0x80
					_, err := conn.Write(response)
					if err != nil {
						fmt.Printf("[ERROR] failed to write response: %v\n", err)
						break
					}
					continue
				}

				start_addr := binary.BigEndian.Uint16(request[8:10])
				qty := binary.BigEndian.Uint16(request[10:12])
				max_addr := int(start_addr) + int(qty) - 1 // we -1 because register is 1-indexing but go is 0-indexing
				if len(device.holding_registers) < max_addr {
					response := append(request[0:8], ERR_ILLEGAL_DATA_ADDRESS)
					response[7] += 0x80
					_, err := conn.Write(response)
					if err != nil {
						fmt.Printf("[ERROR] failed to write response: %v\n", err)
						break
					}
					continue
				}
				var response = make([]byte, 9+(2*qty))
				binary.BigEndian.PutUint16(response[0:2], tx_id)
				binary.BigEndian.PutUint16(response[2:4], protocol_id)
				binary.BigEndian.PutUint16(response[4:6], 2+(qty*2))
				response[6] = slave_id
				response[7] = function_code
				response[8] = byte(qty) * 2

				for addr := 0; addr < int(qty); addr++ {
					idx := addr * 2
					binary.BigEndian.PutUint16(response[(9+idx):], device.holding_registers[addr])
				}
				_, err := conn.Write(response)
				if err != nil {
					fmt.Printf("[ERROR] failed to write response: %v\n", err)
					break
				}
			case READ_COIL:
			default:
				response := append(request[0:8], ERR_ILLEGAL_FUNCTION)
				response[7] += 0x80
				_, err := conn.Write(response)
				if err != nil {
					fmt.Printf("[ERROR] failed to write response: %v\n", err)
					break
				}
			}
		}
	}
}
