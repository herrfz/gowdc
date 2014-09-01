package main

import (
	"flag"
	"github.com/herrfz/gowdc/listeners"
	zmq "github.com/pebbe/zmq4"
)

const (
	HOST      = "localhost"
	INTERFACE = "eth0"
)

func main() {
	port := flag.String("port", "33400", "TCP port to listen")
	flag.Parse()

	c_sock, _ := zmq.NewSocket(zmq.REQ)
	defer c_sock.Close()
	c_sock.Connect("tcp://localhost:5555")

	d_dl_sock, _ := zmq.NewSocket(zmq.PUSH)
	defer d_dl_sock.Close()
	d_dl_sock.Bind("tcp://*:5556")

	d_ul_sock, _ := zmq.NewSocket(zmq.PULL)
	defer d_ul_sock.Close()
	d_ul_sock.Connect("tcp://localhost:5557")

	// Handle connections in a goroutine
	go listeners.ListenTCP(HOST, *port, INTERFACE,
		c_sock, d_dl_sock, d_ul_sock)
	select {}
}
