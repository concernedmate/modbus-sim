package main

import (
	"context"
	"fmt"
	"log"
	"modbus-sim/modbus"
	"os"
	"strconv"
)

func args() {
	if len(os.Args) < 2 {
		fmt.Println("usage:")
		fmt.Println("app modsim")
		fmt.Println("\t start modbus simulator")
		fmt.Println("app modscan ip:port slave_id addr count")
		fmt.Println("\t start modbus scanner")
		fmt.Println("\t example: app modscan 127.0.0.1:502 1 40001 10")
		os.Exit(0)
	}
	switch os.Args[1] {
	case "modsim":
		if err := modbus.ServeTCP("127.0.0.1:8000"); err != nil {
			log.Fatal(err)
		}
	case "modscan":
		if len(os.Args) < 6 {
			fmt.Println("app modscan ip:port slave_id addr count")
			fmt.Println("\t start modbus scanner")
			fmt.Println("\t example: app modscan 127.0.0.1:502 1 40001 10")
			os.Exit(0)
		}
		addr := os.Args[2]
		slave_id, err := strconv.Atoi(os.Args[3])
		if err != nil {
			log.Fatal(fmt.Errorf("invalid slave_id: %v", err))
		}
		register, err := strconv.Atoi(os.Args[4])
		if err != nil {
			log.Fatal(fmt.Errorf("invalid register: %v", err))
		}
		qty, err := strconv.Atoi(os.Args[5])
		if err != nil {
			log.Fatal(fmt.Errorf("invalid qty: %v", err))
		}
		if err := modbus.ConnectTCP(context.Background(), modbus.READ_HOLDING_REGISTERS, addr, uint8(slave_id), register, uint16(qty)); err != nil {
			log.Fatal(err)
		}
	}

}

func main() {
	args()
}
