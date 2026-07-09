package main

import (
	"context"
	"fmt"
	"log"
	"modbus-sim/modbus"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("usage:\n\tmodscan ip:port slave_id addr count\n\tmodscan 127.0.0.1:502 1 40001 10")
		os.Exit(1)
	}
	addr := os.Args[1]
	slave_id, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Fatal(fmt.Errorf("invalid slave_id: %v", err))
	}
	register, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Fatal(fmt.Errorf("invalid register: %v", err))
	}
	qty, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Fatal(fmt.Errorf("invalid qty: %v", err))
	}

	if err := modbus.ConnectTCP(context.Background(), modbus.READ_HOLDING_REGISTERS, addr, uint8(slave_id), register, uint16(qty)); err != nil {
		log.Fatal(err)
	}
}
