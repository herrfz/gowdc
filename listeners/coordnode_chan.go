// This package implements WDC comm listeners
package listeners

import (
	"fmt"
	"net"
)

func ListenCoordNode(d_ul_chan chan []byte, u_conn *net.UDPConn) {
	fmt.Println("Listening to coordinator node uplink channel")

	for {
		cn_buf, more := <-d_ul_chan

		if more {

			// TODO uplink crypto stuffs will be done here

			u_conn.Write(cn_buf)
		} else {
			break
		}

	}
}
