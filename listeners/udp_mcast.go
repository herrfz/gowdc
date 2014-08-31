// This package implements WDC comm listeners
package listeners

import (
	"code.google.com/p/go.net/ipv4"
	"fmt"
	"github.com/herrfz/gowdc/utils"
	zmq "github.com/pebbe/zmq4"
	"net"
	"os"
	"strings"
)

type UMSocket struct {
	socket *ipv4.PacketConn
	group  net.IP
}

func (sock UMSocket) Read() ([]byte, error) {
	buf := make([]byte, 1024)
	// read incoming data into the buffer
	// this blocks until some data are actually received
	dlen, cm, _, err := sock.socket.ReadFrom(buf)
	if err != nil {
		if strings.Contains(err.Error(), "use of closed network connection") {
			return nil, fmt.Errorf("PANIC")
		} else {
			return nil, err
		}
	}

	// process data that are sent to group
	if cm.Dst.IsMulticast() && cm.Dst.Equal(sock.group) {
		// TODO test this if-block
		if len(buf) == 0 || (int(buf[0])+1) != dlen {
			return nil, fmt.Errorf("Error: Inconsistent message length")
		} else {
			return buf[:dlen], nil
		}

	} else {
		return nil, fmt.Errorf("unknown group / not udp mcast")
	}
}

func (sock UMSocket) ReadSerial() ([]byte, error) {
	return nil, nil
}

// args:
// - addr: multicast group/address to listen to
// - port: port number; addr:port builds the mcast socket
// - iface: name of network interface to listen to
// - d_dl_sock:
// - stopch:
func ListenUDPMcast(addr, port, iface string, d_dl_sock *zmq.Socket,
	stopch chan bool) {
	eth, err := net.InterfaceByName(iface)
	if err != nil {
		fmt.Println("Error interface:", err.Error())
		os.Exit(1)
	}

	group := net.ParseIP(addr)
	if group == nil {
		fmt.Println("Error: invalid group address:", addr)
		os.Exit(1)
	}

	// listen to all udp packets on mcast port
	c, err := net.ListenPacket("udp4", "0.0.0.0:"+port)
	if err != nil {
		fmt.Println("Error listening for mcast:", err.Error())
		os.Exit(1)
	}
	// close the listener when the application closes
	defer c.Close()

	// join mcast group
	p := ipv4.NewPacketConn(c)
	if err := p.JoinGroup(eth, &net.UDPAddr{IP: group}); err != nil {
		fmt.Println("Error joining:", err.Error())
		os.Exit(1)
	}
	fmt.Println("Listening on " + addr + ":" + port)

	// enable transmissons of control message
	if err := p.SetControlMessage(ipv4.FlagDst, true); err != nil {
		fmt.Println("Error control message", err.Error())
	}

	c1 := utils.MakeChannel(UMSocket{p, group})

LOOP:
	for {
		select {
		case v1 := <-c1:
			fmt.Println("received UDP multicast")
			// forward to coord node
			d_dl_sock.Send(string(v1), 0)

		case <-stopch:
			break LOOP
		}

	}
}
