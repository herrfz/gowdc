// This package implements WDC comm listeners
package listeners

import (
	"bytes"
	"encoding/binary"
	"fmt"
	msg "github.com/herrfz/gowdc/messages"
	"net"
	"os"
	"strings"
)

const (
	INTERFACE = "eth0"
)

// The listeners
func ListenTCP(host, tcp_port string, dl_chan, ul_chan chan []byte) {
	// WDC state
	connected := 0

	// Listen for TCP incoming connections
	t, err := net.Listen("tcp", host+":"+tcp_port)
	if err != nil {
		fmt.Println("Error listening: ", err.Error())
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
			fmt.Println("Error accepting: ", err.Error())
			continue
		}

		for {
			// Read the incoming data from accepted conn into the buffer
			// this blocks until some data are actually received
			dlen, err := t_conn.Read(buf)
			if err != nil {
				if err.Error() == "EOF" {
					os.Exit(0)
				} else {
					fmt.Println("Error reading: ", err.Error())
					os.Exit(1)
				}
			}

			if (int(buf[0]) + 1) != dlen {
				fmt.Println("Error: Inconsistent message length")
				continue
			}

			switch buf[1] {
			case 0x01: // WDC_CONNECTION_REQ
				// TODO: check expected len of WDC_CONNECTION_REQ
				if dlen < 10 { // <-- this is currently arbitrary
					fmt.Println("Error: wrong WDC_CONNECTION_REQ len")
					continue
				}

				if connected == 1 {
					msg.WDC_ERROR[2] = byte(msg.BUSY_CONNECTED)
					_, err := t_conn.Write(msg.WDC_ERROR)
                    if err != nil {
                        fmt.Println("Error sending WDC_ERROR: ",
                            err.Error())
                    }
					continue
				}

				msg.WDC_GET_STATUS_RES[0] = byte(dlen)
				copy(msg.WDC_GET_STATUS_RES[3:], buf[2:])

				// send message to CoordNode and get a return
                dl_chan <- buf[:dlen]
                cn_buf := <- ul_chan
                
                if int(cn_buf[0] + 1) != len(cn_buf) ||
                cn_buf[1] != 0x02 {  // WDC_CONNECTION_RES
                    fmt.Println("Error on reading from CoordNode")
                    msg.WDC_ERROR[2] = byte(msg.SERIAL_PORT)
                    _, err := t_conn.Write(msg.WDC_ERROR)
                    if err != nil {
                        fmt.Println("Error sending WDC_ERROR: ",
                            err.Error())
                    }
                    continue
                }
                copy(msg.WDC_CONNECTION_RES, cn_buf)

				// get multicast info
				var MCAST_PORT uint16
				temp_buf := bytes.NewReader(buf[8:10])
				err := binary.Read(temp_buf, binary.LittleEndian,
					&MCAST_PORT)
				if err != nil {
					fmt.Println("binary.Read failed: ", err.Error())
				}
				MCAST_ADDR := buf[10 : dlen-1]

				// create UDP socket to server
				SERVER_SOCK := t_conn.RemoteAddr().String()
				SERVER_IP := strings.Split(SERVER_SOCK, ":")[0]
				var SERVER_UDP_PORT uint16
				temp_buf = bytes.NewReader(buf[6:8])
				err = binary.Read(temp_buf, binary.LittleEndian,
					&SERVER_UDP_PORT)
				// create UDP address from string
				udpaddr, err := net.ResolveUDPAddr("udp",
					SERVER_IP+":"+fmt.Sprintf("%d", SERVER_UDP_PORT))
				if err != nil {
					fmt.Println("Error resolving UDP: ", err.Error())
					continue
				}
				// dial UDP
				u_conn, err := net.DialUDP("udp", nil, udpaddr)
				if err != nil {
					fmt.Println("Error server UDP: ", err.Error())
					msg.WDC_ERROR[2] = byte(msg.CONNECTING)
					_, err := t_conn.Write(msg.WDC_ERROR)
                    if err != nil {
                        fmt.Println("Error sending WDC_ERROR: ",
                            err.Error())
                    }
					continue
				}
				fmt.Printf("Server UDP is at: %s:%d\n",
                    SERVER_IP, SERVER_UDP_PORT)
				// close UDP when the application closes
				defer u_conn.Close()

				// TODO, this is a  fake 8 Byte coordnode long address
				// should be taken from serial read
				fake := []byte{0xde, 0xad, 0xbe, 0xef,
					0xde, 0xad, 0xbe, 0xef}
				copy(msg.WDC_CONNECTION_RES[2:], fake)
				// reply to server
				_, err = t_conn.Write(msg.WDC_CONNECTION_RES)
                if err != nil {
                    fmt.Println("Error sending WDC_CONNECTION_RES: ",
                        err.Error())
                }

				// Serve UDP mcast in a new goroutine
				go ListenUDPMcast(string(MCAST_ADDR),
					fmt.Sprintf("%d", MCAST_PORT), INTERFACE, u_conn,
                    dl_chan, ul_chan)

				connected = 1

			case 0x03: // WDC_DISCONNECTION_REQ
				if connected == 1 {
					// TODO stop listening UDP mcast
					// TODO u_conn.Close() (u_conn undefined)

                    // send disconnect to CoordNode (bye)
                    dl_chan <- buf[:dlen]
                    msg.WDC_DISCONNECTION_REQ_ACK = <- ul_chan
					_, err := t_conn.Write(
                        msg.WDC_DISCONNECTION_REQ_ACK)
                    if err != nil {
                        fmt.Println("Error sending WDC_DISC_REQ_ACK: ",
                            err.Error())
                    }
				}

				connected = 0

			case 0x05: // WDC_GET_STATUS_REQ
				msg.WDC_GET_STATUS_RES[2] = byte(connected)
				_, err := t_conn.Write(msg.WDC_GET_STATUS_RES)
                if err != nil {
                    fmt.Println("Error sending WDC_GET_STATUS_RES: ",
                        err.Error())
                }

			case 0x07, 0x09:
				// WDC_SET_COOR_LONG_ADDR_REQ ||
				// WDC_RESET_REQ
				if connected == 1 {
					msg.WDC_ERROR[2] = byte(msg.BUSY_CONNECTED)
					_, err := t_conn.Write(msg.WDC_ERROR)
                    if err != nil {
                    fmt.Println("Error sending WDC_ERROR: ",
                        err.Error())
                    }

				} else {

					if buf[1] == 0x07 {
                        dl_chan <- buf[:dlen]
                        msg.WDC_SET_COOR_LONG_ADDR_REQ_ACK = <- ul_chan
						_, err := t_conn.Write(msg.WDC_SET_COOR_LONG_ADDR_REQ_ACK)
                        if err != nil {
                            fmt.Println("Error send LONG_ADDR_ACK: ",
                                err.Error())
                        }

					} else {
						_, err := t_conn.Write(msg.WDC_RESET_REQ_ACK)
                        if err != nil {
                            fmt.Println("Error send RESET_ACK: ",
                                err.Error())
                        }
						// Exit will close UDP, TCP, UDP mcast
						// TODO: serial; reboot instead of app exit
						os.Exit(0)
					}
				}

			default:
				msg.WDC_ERROR[2] = byte(msg.WRONG_CMD)
				_, err := t_conn.Write(msg.WDC_ERROR)
                if err != nil {
                    fmt.Println("Error sending WDC_ERROR: ",
                        err.Error())
                }
			}
		}

        // TODO send WDC_DISCONNECTION_REQ to CoordNode
	}
}
