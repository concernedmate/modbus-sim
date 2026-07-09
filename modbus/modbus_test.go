package modbus_test

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"modbus-sim/modbus"
	"strings"
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

func Test_ParseResponseFC03TCP(t *testing.T) {
	var data_count uint8 = 127 // max 127
	var byte_count uint8 = uint8(data_count) * 2

	var data strings.Builder
	for idx := range data_count {
		fmt.Fprintf(&data, "%04s ", hex.EncodeToString([]byte{uint8(idx)}))
	}

	hex_data := strings.ReplaceAll(fmt.Sprintf("0001 0000 0009 0103 %02s %s", hex.EncodeToString([]byte{byte_count}), data.String()), " ", "")
	resp, err := hex.DecodeString(hex_data)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	result, err := modbus.ParseResponseFC03TCP(resp, 0, uint16(data_count))
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	for key, val := range result {
		if key != int(binary.BigEndian.Uint16(val)) {
			t.Errorf("expected %d, got %d", key, int(binary.BigEndian.Uint16(val)))
		}
	}

	hex_data = strings.ReplaceAll(fmt.Sprintf("0001 0000 0009 0101 %02s %s", hex.EncodeToString([]byte{byte_count}), data.String()), " ", "")
	resp, err = hex.DecodeString(hex_data)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	result, err = modbus.ParseResponseFC03TCP(resp, 0, uint16(data_count))
	if err == nil {
		t.Errorf("should fail")
	}

	error_codes := map[string]int{
		"exception: ILLEGAL_FUNCTION":                        modbus.ERR_ILLEGAL_FUNCTION,
		"exception: ILLEGAL_DATA_ADDRESS":                    modbus.ERR_ILLEGAL_DATA_ADDRESS,
		"exception: ILLEGAL_DATA_VALUE":                      modbus.ERR_ILLEGAL_DATA_VALUE,
		"exception: SLAVE_DEVICE_FAILURE":                    modbus.ERR_SLAVE_DEVICE_FAILURE,
		"exception: ACKNOWLEDGE":                             modbus.ERR_ACKNOWLEDGE,
		"exception: SLAVE_DEVICE_BUSY":                       modbus.ERR_SLAVE_DEVICE_BUSY,
		"exception: NEGATIVE_ACKNOWLEDGE":                    modbus.ERR_NEGATIVE_ACKNOWLEDGE,
		"exception: MEMORY_PARITY_ERROR":                     modbus.ERR_MEMORY_PARITY_ERROR,
		"exception: GATEWAY_PATH_UNAVAILABLE":                modbus.ERR_GATEWAY_PATH_UNAVAILABLE,
		"exception: GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND": modbus.ERR_GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND,
	}

	for key, val := range error_codes {
		resp, err = hex.DecodeString(fmt.Sprintf("0001000000090181%02s", hex.EncodeToString([]byte{uint8(val)})))
		if err != nil {
			t.Errorf("should not fail: %v", err)
		}
		result, err = modbus.ParseResponseFC03TCP(resp, 0, uint16(data_count))
		if err == nil {
			t.Errorf("should fail")
		}
		if err.Error() != key {
			t.Errorf("expected %s, got %s", key, err.Error())
		}
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
