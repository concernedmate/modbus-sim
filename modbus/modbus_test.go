package modbus_test

import (
	"encoding/hex"
	"modbus-sim/modbus"
	"testing"
)

func Test_FrameFC03TCP(t *testing.T) {
	register, err := modbus.ParseRegisterAddr(40108)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	if register != 107 {
		t.Errorf("expected %d, got %d", 107, register)
	}

	result := modbus.FrameFC03TCP(1, 17, register, 3)
	if hex.EncodeToString(result) != "0001000000061103006b0003" {
		t.Errorf("expected %s, got %s", "0001000000061103006b0003", hex.EncodeToString(result))
	}
}

func Test_FrameFC03RTU(t *testing.T) {
	register, err := modbus.ParseRegisterAddr(40108)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	if register != 107 {
		t.Errorf("expected %d, got %d", 107, register)
	}

	result := modbus.FrameFC03RTU(17, register, 3)
	if hex.EncodeToString(result) != "1103006b00037687" {
		t.Errorf("expected %s, got %s", "1103006b00037687", hex.EncodeToString(result))
	}
}
