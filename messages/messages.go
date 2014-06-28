// This package maintains protocol messages and error codes
package messages

//
// Messages
//
var (
	WDC_ERROR                      = make([]byte, 3)
	WDC_CONNECTION_RES             = make([]byte, 10)
	WDC_DISCONNECTION_REQ_ACK      = make([]byte, 2)
	WDC_GET_STATUS_RES             = make([]byte, 64)
	WDC_SET_COOR_LONG_ADDR_REQ_ACK = make([]byte, 2)
	WDC_RESET_REQ_ACK              = make([]byte, 2)
	WDC_ACK                        = make([]byte, 2)
	WDC_GET_TDMA_RES               = make([]byte, 24)
)

//
// Error codes
//
var (
	SERIAL_PORT    = 0x00
	BUSY_CONNECTED = 0x01
	CONNECTING     = 0x02
	WRONG_CMD      = 0x03
)

func init() {
	copy(WDC_ERROR[:], []byte{2, 0x00})
	copy(WDC_CONNECTION_RES[:], []byte{9, 0x02})
	copy(WDC_DISCONNECTION_REQ_ACK[:], []byte{1, 0x04})
	copy(WDC_GET_STATUS_RES[:], []byte{10, 0x06})
	copy(WDC_SET_COOR_LONG_ADDR_REQ_ACK[:], []byte{1, 0x08})
	copy(WDC_RESET_REQ_ACK[:], []byte{1, 0x0a})
	copy(WDC_ACK[:], []byte{1})
	copy(WDC_GET_TDMA_RES[:], []byte{23, 0x16})
}
