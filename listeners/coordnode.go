// This package implements WDC comm listeners
package listeners

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
	"net"
)

func ListenCoordNode(ctrl chan int, d_ul_sock *zmq.Socket, u_conn *net.UDPConn) {
	fmt.Println("Listening to coordinator node uplink")

	for {
		//select {
		//case <-ctrl:
		//	break
		//default:
		cn_buf, _ := d_ul_sock.Recv(0)
		fmt.Println("received from coordinator node")
		u_conn.Write([]byte(cn_buf))
		fmt.Println("sent over UDP")
		//}
	}
}
