package modbus

// function codes
const (
	READ_COIL              uint8 = 0x01
	READ_DISCRETE_INPUT    uint8 = 0x02
	READ_HOLDING_REGISTERS uint8 = 0x03
	READ_INPUT_REGISTERS   uint8 = 0x04

	WRITE_SINGLE_COIL                uint8 = 0x05
	WRITE_SINGLE_HOLDING_REGISTERS   uint8 = 0x06
	WRITE_MULTIPLE_COILS             uint8 = 0x15
	WRITE_MULTIPLE_HOLDING_REGISTERS uint8 = 0x16
)

// exception codes
const (
	ERR_ILLEGAL_FUNCTION                        uint8 = 0x01
	ERR_ILLEGAL_DATA_ADDRESS                    uint8 = 0x02
	ERR_ILLEGAL_DATA_VALUE                      uint8 = 0x03
	ERR_SLAVE_DEVICE_FAILURE                    uint8 = 0x04
	ERR_ACKNOWLEDGE                             uint8 = 0x05
	ERR_SLAVE_DEVICE_BUSY                       uint8 = 0x06
	ERR_NEGATIVE_ACKNOWLEDGE                    uint8 = 0x07
	ERR_MEMORY_PARITY_ERROR                     uint8 = 0x08
	ERR_GATEWAY_PATH_UNAVAILABLE                uint8 = 0x10
	ERR_GATEWAY_TARGET_DEVICE_FAILED_TO_RESPOND uint8 = 0x11
)

// protocol
const (
	RTU = "RTU"
	TCP = "TCP"
)

var modbus_exception = map[uint8]string{
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
