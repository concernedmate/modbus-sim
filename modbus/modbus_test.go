package modbus_test

import (
	"encoding/hex"
	"modbus-sim/modbus"
	"testing"
)

func Test_FrameFC03TCP(t *testing.T) {
	result := modbus.FrameFC03TCP(1, 17, modbus.ConvertRegisterAddr(40108), 3)
	if hex.EncodeToString(result) != "0001000000061103006b0003" {
		t.Errorf("expected %s, got %s", "0001000000061103006b0003", hex.EncodeToString(result))
	}

	result = modbus.FrameFC03TCP(1, 17, 107, 3)
	if hex.EncodeToString(result) != "0001000000061103006b0003" {
		t.Errorf("expected %s, got %s", "0001000000061103006b0003", hex.EncodeToString(result))
	}
}

func Test_FrameFC03RTU(t *testing.T) {
	result := modbus.FrameFC03RTU(17, modbus.ConvertRegisterAddr(40108), 3)
	if hex.EncodeToString(result) != "1103006b00037687" {
		t.Errorf("expected %s, got %s", "1103006b00037687", hex.EncodeToString(result))
	}

	result = modbus.FrameFC03RTU(17, 107, 3)
	if hex.EncodeToString(result) != "1103006b00037687" {
		t.Errorf("expected %s, got %s", "1103006b00037687", hex.EncodeToString(result))
	}
}
