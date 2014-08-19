// This package implements WDC comm listeners
package listeners

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	msg "github.com/herrfz/gowdc/messages"
	zmq "github.com/pebbe/zmq4"
	"net"
	"os"
	"strings"
)

// args:
// - host: hostname
// - tcp_port: port number; host:port builds the tcp socket
// - iface: name of network interface to listen to
// - c_sock:
// - d_dl_sock:
// - d_ul_sock:
func ListenTCP(host, tcp_port, iface string,
	c_sock, d_dl_sock, d_ul_sock *zmq.Socket) {
	// WDC state
	connected := 0

	// control channel to stop listening to coordnode
	cn_stopch := make(chan bool)
	// control channel to stop listening udp mcast
	mcast_stopch := make(chan bool)

	// Listen for TCP incoming connections
	t, err := net.Listen("tcp", host+":"+tcp_port)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes
	defer t.Close()

	// Make a buffer to hold incoming data
	buf := make([]byte, 1024)
	for {
		// Listen for an incoming TCP connection
		fmt.Println("Listening on " + host + ":" + tcp_port)
		// this blocks until someone attempts to connect
		t_conn, err := t.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err.Error())
			continue
		}
		defer t_conn.Close()

		for {
			// Read the incoming data from accepted conn into the buffer
			// this blocks until some data are actually received
			dlen, err := t_conn.Read(buf)
			if err != nil {
				if err.Error() == "EOF" {
					if connected == 1 {
						// stop listening to CoordNode
						cn_stopch <- true
						// stop listening UDP mcast
						mcast_stopch <- true
						connected = 0
					}
					break
				} else {
					fmt.Println("Error reading:", err.Error())
					continue
				}
			}

			if len(buf) == 0 || (int(buf[0])+1) != dlen {
				fmt.Println("Error: Inconsistent message length")
				continue
			}

			fmt.Println("received TCP command",
				hex.EncodeToString(buf[:dlen]))

			switch buf[1] {
			case 0x01: // WDC_CONNECTION_REQ
				fmt.Println("received connection request:",
					hex.EncodeToString(buf[:dlen]))
				// TODO: check expected len of WDC_CONNECTION_REQ
				if dlen < 10 { // <-- this is currently arbitrary
					fmt.Println("Error: wrong WDC_CONNECTION_REQ len")
					continue
				}

				if connected == 1 {
					msg.WDC_ERROR[2] = byte(msg.BUSY_CONNECTED)
					t_conn.Write(msg.WDC_ERROR)
					continue
				}

				msg.WDC_GET_STATUS_RES[0] = byte(dlen)
				copy(msg.WDC_GET_STATUS_RES[3:], buf[2:])
				// orig msg is very long with trailing zeroes
				// truncate to dlen
				msg.WDC_GET_STATUS_RES = msg.WDC_GET_STATUS_RES[:dlen+1]

				// send message to CoordNode and get a return
				c_sock.Send(string(buf[:dlen]), 0)
				cn_buf, _ := c_sock.Recv(0)

				if int(cn_buf[0])+1 != len(cn_buf) ||
					cn_buf[1] != 0x02 { // WDC_CONNECTION_RES
					fmt.Println("Error on reading from CoordNode")
					msg.WDC_ERROR[2] = byte(msg.SERIAL_PORT)
					t_conn.Write(msg.WDC_ERROR)
					continue
				}

				// get multicast info
				var MCAST_PORT uint16
				binary.Read(bytes.NewReader(buf[8:10]),
					binary.LittleEndian, &MCAST_PORT)
				MCAST_ADDR := buf[10 : dlen-1]

				// create UDP socket to server
				SERVER_SOCK := t_conn.RemoteAddr().String()
				SERVER_IP := strings.Split(SERVER_SOCK, ":")[0]
				var SERVER_UDP_PORT uint16
				binary.Read(bytes.NewReader(buf[6:8]),
					binary.LittleEndian, &SERVER_UDP_PORT)
				// create UDP address from string
				udpaddr, err := net.ResolveUDPAddr("udp",
					SERVER_IP+":"+fmt.Sprintf("%d", SERVER_UDP_PORT))
				if err != nil {
					fmt.Println("Error resolving UDP:", err.Error())
					continue
				}
				// dial UDP
				u_conn, err := net.DialUDP("udp", nil, udpaddr)
				if err != nil {
					fmt.Println("Error server UDP:", err.Error())
					msg.WDC_ERROR[2] = byte(msg.CONNECTING)
					t_conn.Write(msg.WDC_ERROR)
					continue
				}
				fmt.Printf("Server UDP is at: %s:%d\n",
					SERVER_IP, SERVER_UDP_PORT)
				// close UDP when the application closes
				defer u_conn.Close()

				// set and send connection response
				copy(msg.WDC_CONNECTION_RES, cn_buf)
				// reply to server
				t_conn.Write(msg.WDC_CONNECTION_RES)
				fmt.Println("sent connection response:",
					hex.EncodeToString(msg.WDC_CONNECTION_RES))

				// Serve UDP mcast in a new goroutine
				go ListenUDPMcast(string(MCAST_ADDR),
					fmt.Sprintf("%d", MCAST_PORT), iface, d_dl_sock,
					mcast_stopch)

				// Start listening to CoordNode
				go ListenCoordNode(d_ul_sock, u_conn, cn_stopch)

				connected = 1

			case 0x03: // WDC_DISCONNECTION_REQ
				fmt.Println("received disconnection request")
				if connected == 1 {
					// send disconnect to CoordNode (bye)
					c_sock.Send(string(buf[:dlen]), 0)
					req_ack, _ := c_sock.Recv(0)
					msg.WDC_DISCONNECTION_REQ_ACK = []byte(req_ack)

					// stop listening to CoordNode
					cn_stopch <- true
					// stop listening UDP mcast
					mcast_stopch <- true

					// TODO u_conn.Close() (u_conn undefined)

					// send disconnect ack to server
					t_conn.Write(msg.WDC_DISCONNECTION_REQ_ACK)
					fmt.Println("sent disconnection request ack, bye!")
				}

				connected = 0

			case 0x05: // WDC_GET_STATUS_REQ
				fmt.Println("received WDC status request")
				msg.WDC_GET_STATUS_RES[2] = byte(connected)
				t_conn.Write(msg.WDC_GET_STATUS_RES)
				fmt.Println("sent WDC status response:",
					hex.EncodeToString(msg.WDC_GET_STATUS_RES))

			case 0x07, 0x09:
				// WDC_SET_COOR_LONG_ADDR_REQ ||
				// WDC_RESET_REQ
				if connected == 1 {
					msg.WDC_ERROR[2] = byte(msg.BUSY_CONNECTED)
					t_conn.Write(msg.WDC_ERROR)

				} else {

					fmt.Println("received long address set req")

					c_sock.Send(string(buf[:dlen]), 0)
					cn_buf, _ := c_sock.Recv(0)

					if int(cn_buf[0])+1 != len(cn_buf) {
						fmt.Println("Error reading from CoordNode")
						continue
					}

					copy(msg.WDC_SET_COOR_LONG_ADDR_REQ_ACK, cn_buf)
					t_conn.Write(msg.WDC_SET_COOR_LONG_ADDR_REQ_ACK)
					fmt.Println("sent long address set ack")

					if buf[1] == 0x09 {
						fmt.Println("received reset request")
						t_conn.Write(msg.WDC_RESET_REQ_ACK)
						fmt.Println("sent reset ack, bye!")
						// Exit will close UDP, TCP, UDP mcast
						// TODO: serial; reboot instead of app exit
						os.Exit(0)
					}
				}

			default:
				fmt.Println("received unknown command")
				msg.WDC_ERROR[2] = byte(msg.WRONG_CMD)
				t_conn.Write(msg.WDC_ERROR)
			}
		}
		// TODO cleanup on exiting loop
	}
}
