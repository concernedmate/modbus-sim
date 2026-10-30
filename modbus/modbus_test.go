package modbus_test

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"modbus-sim/modbus"
	"strings"
	"testing"
)

func Test_RequestFrameTCP(t *testing.T) {
	register, err := modbus.ParseRegisterAddr(40108)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	if register != 107 {
		t.Errorf("expected %d, got %d", 107, register)
	}

	valid := map[uint8]struct{}{
		modbus.READ_COIL:              {},
		modbus.READ_DISCRETE_INPUT:    {},
		modbus.READ_HOLDING_REGISTERS: {},
		modbus.READ_INPUT_REGISTERS:   {},
	}
	for fc := range valid {
		result, err := modbus.RequestFrameTCP(fc, 1, 17, register, 3)
		if err != nil {
			t.Errorf("should not fail: %v", err)
		}

		frame := fmt.Sprintf("00010000000611%02d006b0003", fc)
		if hex.EncodeToString(result) != frame {
			t.Errorf("expected %s, got %s", frame, hex.EncodeToString(result))
		}
	}
	for fc := range math.MaxUint8 + 1 {
		if _, ok := valid[uint8(fc)]; ok {
			continue
		}
		_, err := modbus.RequestFrameTCP(uint8(fc), 1, 17, register, 3)
		if err == nil {
			t.Errorf("should fail on fc %02d", fc)
		}
	}
}

func Test_RequestFrameRTU(t *testing.T) {
	register, err := modbus.ParseRegisterAddr(40108)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	if register != 107 {
		t.Errorf("expected %d, got %d", 107, register)
	}

	valid := map[uint8]struct{}{
		modbus.READ_COIL:              {},
		modbus.READ_DISCRETE_INPUT:    {},
		modbus.READ_HOLDING_REGISTERS: {},
		modbus.READ_INPUT_REGISTERS:   {},
	}
	for fc := range valid {
		result, err := modbus.RequestFrameRTU(fc, 17, register, 3)
		if err != nil {
			t.Errorf("should not fail: %v", err)
		}

		frame_nocrc := fmt.Sprintf("11%02d006b0003", fc)
		if hex.EncodeToString(result[:len(result)-2]) != frame_nocrc {
			t.Errorf("expected %s, got %s", frame_nocrc, hex.EncodeToString(result))
		}
	}
	for fc := range math.MaxUint8 + 1 {
		if _, ok := valid[uint8(fc)]; ok {
			continue
		}
		_, err := modbus.RequestFrameRTU(uint8(fc), 17, register, 3)
		if err == nil {
			t.Errorf("should fail on fc %02d", fc)
		}
	}
}

func Test_ParseResponseTCP(t *testing.T) {
	var data_count uint8 = 127 // max 127
	var byte_count uint8 = uint8(data_count) * 2

	var data strings.Builder
	for idx := range data_count {
		fmt.Fprintf(&data, "%04s ", hex.EncodeToString([]byte{uint8(idx)}))
	}

	// test s
	hex_data := strings.ReplaceAll(fmt.Sprintf("0001 0000 0009 01 03 %02s %s", hex.EncodeToString([]byte{byte_count}), data.String()), " ", "")
	resp, err := hex.DecodeString(hex_data)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	result, err := modbus.ParseResponseTCP(resp, 0, uint16(data_count))
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	if len(result) != int(data_count) {
		t.Errorf("result should have %d data, got %d", data_count, len(result))
	}
	for key, val := range result {
		if key != int(binary.BigEndian.Uint16(val)) {
			t.Errorf("expected %d, got %d", key, int(binary.BigEndian.Uint16(val)))
		}
	}

	// test different length
	hex_data = strings.ReplaceAll(fmt.Sprintf("0001 0000 0009 01 03 %02s %s", hex.EncodeToString([]byte{byte_count - 1}), data.String()), " ", "")
	resp, err = hex.DecodeString(hex_data)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	result, err = modbus.ParseResponseTCP(resp, 0, uint16(data_count))
	if err == nil {
		t.Errorf("should fail")
	}

	// test error_codes
	error_codes := map[string]uint8{
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
		hex_data := strings.ReplaceAll(fmt.Sprintf("0001 0000 0009 01 81 %02s", hex.EncodeToString([]byte{uint8(val)})), " ", "")
		resp, err = hex.DecodeString(hex_data)
		if err != nil {
			t.Errorf("should not fail: %v", err)
		}
		result, err = modbus.ParseResponseTCP(resp, 0, uint16(data_count))
		if err == nil {
			t.Errorf("should fail")
		}
		if err.Error() != key {
			t.Errorf("expected %s, got %s", key, err.Error())
		}
	}
}

func Test_ParseResponseRTU(t *testing.T) {
	var data_count uint8 = 127 // max 127
	var byte_count uint8 = uint8(data_count) * 2

	var data strings.Builder
	for idx := range data_count {
		fmt.Fprintf(&data, "%04s ", hex.EncodeToString([]byte{uint8(idx)}))
	}

	// test success
	hex_data := strings.ReplaceAll(fmt.Sprintf("01 03 %02s %s", hex.EncodeToString([]byte{byte_count}), data.String()), " ", "")
	resp, err := hex.DecodeString(hex_data)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	resp = append(resp, modbus.CalculateCRCFast(resp)...)
	result, err := modbus.ParseResponseRTU(resp, 0, uint16(data_count))
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	if len(result) != int(data_count) {
		t.Errorf("result should have %d data, got %d", data_count, len(result))
	}
	for key, val := range result {
		if key != int(binary.BigEndian.Uint16(val)) {
			t.Errorf("expected %d, got %d", key, int(binary.BigEndian.Uint16(val)))
		}
	}

	// test different length
	hex_data = strings.ReplaceAll(fmt.Sprintf("01 03 %02s %s", hex.EncodeToString([]byte{byte_count - 1}), data.String()), " ", "")
	resp, err = hex.DecodeString(hex_data)
	if err != nil {
		t.Errorf("should not fail: %v", err)
	}
	resp = append(resp, modbus.CalculateCRCFast(resp)...)
	result, err = modbus.ParseResponseRTU(resp, 0, uint16(data_count))
	if err == nil {
		t.Errorf("should fail")
	}

	// test error_codes
	error_codes := map[string]uint8{
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
		hex_data := strings.ReplaceAll(fmt.Sprintf("01 81 %02s", hex.EncodeToString([]byte{uint8(val)})), " ", "")
		resp, err = hex.DecodeString(hex_data)
		if err != nil {
			t.Errorf("should not fail: %v", err)
		}
		resp = append(resp, modbus.CalculateCRCFast(resp)...)
		result, err = modbus.ParseResponseRTU(resp, 0, uint16(data_count))
		if err == nil {
			t.Errorf("should fail")
		}
		if err.Error() != key {
			t.Errorf("expected %s, got %s", key, err.Error())
		}
	}
}
