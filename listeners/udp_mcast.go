// This package implements WDC comm listeners
package listeners

import (
	"code.google.com/p/go.net/ipv4"
	"fmt"
	msg "github.com/herrfz/gowdc/messages"
	zmq "github.com/pebbe/zmq4"
	"net"
	"os"
)

type MCastObj struct {
	dlen int
	cm   *ipv4.ControlMessage
	err  error
	buf  []byte
}

// args:
// - addr: multicast group/address to listen to
// - port: port number; addr:port builds the mcast socket
// - iface: name of network interface to listen to
// - d_dl_sock:
// - stopch:
func ListenUDPMcast(addr, port, iface string, d_dl_sock *zmq.Socket, stopch chan bool) {
	eth, err := net.InterfaceByName(iface)
	if err != nil {
		fmt.Println("Error interface: ", err.Error())
		os.Exit(1)
	}

	group := net.ParseIP(addr)
	if group == nil {
		fmt.Println("Error: invalid group address")
		os.Exit(1)
	}

	// listen to all udp packets on mcast port
	c, err := net.ListenPacket("udp4", "0.0.0.0:"+port)
	if err != nil {
		fmt.Println("Error listening for mcast: ", err.Error())
		os.Exit(1)
	}
	// close the listener when the application closes
	defer c.Close()

	// join mcast group
	p := ipv4.NewPacketConn(c)
	if err := p.JoinGroup(eth, &net.UDPAddr{IP: group}); err != nil {
		fmt.Println("Error joining: ", err.Error())
		os.Exit(1)
	}
	fmt.Println("Listening on " + addr + ":" + port)

	// enable transmissons of control message
	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
		fmt.Println("Error control message", err.Error())
	}

	c1 := makeChannel(p)
LOOP:
	for {
		select {
		case v1 := <-c1:
			fmt.Println("received UDP multicast")
			if v1.err != nil {
				fmt.Println("Error reading: ", v1.err.Error())
				continue LOOP
			}

			// process data that are sent to group
			if v1.cm.Dst.IsMulticast() && v1.cm.Dst.Equal(group) {
				// TODO test this if-block
				if v1.dlen == 0 || (int(v1.buf[0])+1) != v1.dlen {
					fmt.Println("Error: Inconsistent message length")
					msg.WDC_ERROR[2] = byte(msg.WRONG_CMD)
				}

				d_dl_sock.Send(string(v1.buf[:v1.dlen]), 0)

			} else {
				// unknown group / not udp mcast, discard
				continue LOOP
			}

		case <-stopch:
			break LOOP
		}

	}
}

func makeChannel(p *ipv4.PacketConn) <-chan MCastObj {
	c := make(chan MCastObj)
	buf := make([]byte, 1024)
	go func() {
		for {
			// read incoming data into the buffer
			// this blocks until some data are actually received
			dlen, cm, _, err := p.ReadFrom(buf)
			mcast_data := MCastObj{dlen, cm, err, buf}
			c <- mcast_data
		}
	}()
	return c
}
