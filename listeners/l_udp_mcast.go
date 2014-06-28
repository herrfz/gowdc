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
// - u_conn: UDP return channel
// - dl_chan: channel for sending downlink messages
// - ul_chan: channel for sending uplink messages
func ListenUDPMcast(addr, port, iface string, u_conn *net.UDPConn,
    dl_chan, ul_chan chan []byte) {
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

			switch buf[1] {
			case 0x10: // syn TDMA
				fmt.Println("sync-ing WDC")

			case 0x11: // start TDMA
				fmt.Println("starting TDMA")
				copy(msg.WDC_GET_TDMA_RES[:],
					[]byte{byte(dlen + 1), 0x16, 0x01})
				copy(msg.WDC_GET_TDMA_RES[3:], buf[2:])

				// TODO write serial

				// send back ACK via UDP
                msg.WDC_ACK[1] = 0x12  // START_TDMA_REQ_ACK
                u_conn.Write(msg.WDC_ACK)

			case 0x13: // stop TDMA
				fmt.Println("stopping TDMA")
				// send back ACK
                msg.WDC_ACK[1] = 0x14  // STOP_TDMA_REQ_ACK
                u_conn.Write(msg.WDC_ACK)

			case 0x15: // TDMA status
				fmt.Println("sending TDMA status response")

				// TODO read serial (get status)

				// TODO send back WDC_GET_TDMA_RES via UDP

			case 0x17:  // data request
				fmt.Println("data request")

				// TODO

			default:
				fmt.Println("wrong cmd")
				// send back WDC_ERROR via UDP
                msg.WDC_ERROR[2] = byte(msg.WRONG_CMD)
                u_conn.Write(msg.WDC_ERROR)
			}

		} else {
			// unknown group / not udp mcast, discard
			continue
		}
	}
}
