// This package implements WDC comm listeners
package listeners

import (
	"fmt"
	"github.com/herrfz/gowdc/utils"
	zmq "github.com/pebbe/zmq4"
	"net"
)

type CNSocket struct {
	socket *zmq.Socket
}

func (sock CNSocket) Read() ([]byte, error) {
	cn_buf, err := sock.socket.Recv(0)
	if err != nil {
		// in very very rare cases it crashes
		// ignore for now, until I figure out why
		return nil, fmt.Errorf("DONTPANIC")
	} else {
		return []byte(cn_buf), err
	}
}

func ListenCoordNode(d_ul_sock *zmq.Socket, u_conn *net.UDPConn,
	stopch chan bool) {
	fmt.Println("Listening to coordinator node uplink")

	cn_ch := utils.MakeChannel(CNSocket{d_ul_sock})

LOOP:
	for {
		select {
		case cn_buf := <-cn_ch:
			fmt.Println("received from coordinator node")
			u_conn.Write([]byte(cn_buf))
			fmt.Println("sent coord node response over UDP")

		case <-stopch:
			break LOOP
		}
	}
}
