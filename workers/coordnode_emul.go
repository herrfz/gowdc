// This package implements workers for external components
package workers

import (
    "encoding/hex"
    "fmt"
    msg "github.com/herrfz/gowdc/messages"
)

func EmulCoordNode(dl_chan, c_ul_chan, d_ul_chan chan []byte) {
    tdma_status := msg.WDC_GET_TDMA_RES

	for {
        buf := <-dl_chan
        dlen := len(buf)

		switch buf[1] {
		case 0x01:
            fmt.Println("received CoordNode connect")
			msg.WDC_CONNECTION_RES[2] = 3
            c_ul_chan <- msg.WDC_CONNECTION_RES
            fmt.Println("CoordNode connection created")

		case 0x03:
            fmt.Println("received CoordNode disconnect")
			c_ul_chan <- msg.WDC_DISCONNECTION_REQ_ACK
            fmt.Println("CoordNode disconnected")

		case 0x07:
            fmt.Println("received set CorrdNode long address")
			c_ul_chan <- msg.WDC_SET_COOR_LONG_ADDR_REQ_ACK
            fmt.Println("CorrdNode long address set")

        case 0x10:
            fmt.Println("received WDC sync")  // TODO
            fmt.Println("WDC sync-ed")  // TODO

        case 0x11: // start TDMA
            fmt.Println("received start TDMA")
            copy(tdma_status[:], []byte{byte(dlen), 0x16, 0x01})
            copy(tdma_status[3:], buf[2:])

            msg.WDC_ACK[1] = 0x12 // START_TDMA_REQ_ACK
            d_ul_chan <- msg.WDC_ACK
            fmt.Println("TDMA started")

        case 0x13: // stop TDMA
            fmt.Println("received stop TDMA")
            msg.WDC_ACK[1] = 0x14 // STOP_TDMA_REQ_ACK
            d_ul_chan <- msg.WDC_ACK
            fmt.Println("TDMA stopped")

        case 0x15: // TDMA status
            fmt.Println("received TDMA status request")
            d_ul_chan <- tdma_status
            fmt.Println("sent TDMA status response: ", 
                hex.EncodeToString(tdma_status))

        case 0x17: // data request
            fmt.Println("received data request")

            // TODO

        default:
            fmt.Println("received wrong cmd")
            // send back WDC_ERROR
            msg.WDC_ERROR[2] = byte(msg.WRONG_CMD)
            // TODO confirm this (data, not command)
            d_ul_chan <- msg.WDC_ERROR
		}
	}
}
