// This package implements WDC comm listeners
package listeners

import (
    "fmt"
    "net"
)

func ListenCoordNode(d_ul_chan chan []byte, u_conn *net.UDPConn) {
    fmt.Println("Listening to coordinator node uplink channel")

    for {
        cn_buf := <-d_ul_chan
        u_conn.Write(cn_buf)
    }
}
