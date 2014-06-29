// This package implements WDC comm listeners
package listeners

import (
	"code.google.com/p/go.net/ipv4"
	"fmt"
	msg "github.com/herrfz/gowdc/messages"
	"net"
	"os"
)

// args:
// - addr: multicast group/address to listen to
// - port: port number; addr:port builds the mcast socket
// - iface: name of network interface to listen to
// - dl_chan: channel for sending downlink data
func ListenUDPMcast(addr, port, iface string, dl_chan chan []byte) {
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

	buf := make([]byte, 1024)
	for {
		// read incoming data into the buffer
		// this blocks until some data are actually received
		dlen, cm, _, err := p.ReadFrom(buf)
		if err != nil {
			fmt.Println("Error reading: ", err.Error())
			continue
		}

		// process data that are sent to group
		if cm.Dst.IsMulticast() && cm.Dst.Equal(group) {
			// TODO test this if-block
			if dlen == 0 || (int(buf[0])+1) != dlen {
				fmt.Println("Error: Inconsistent message length")
				msg.WDC_ERROR[2] = byte(msg.WRONG_CMD)
			}

			dl_chan <- buf[:dlen]

		} else {
			// unknown group / not udp mcast, discard
			continue
		}
	}
}
