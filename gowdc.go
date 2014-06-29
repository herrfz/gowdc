package main

import (
	"flag"
	"github.com/herrfz/gowdc/listeners"
	"github.com/herrfz/gowdc/workers"
)

const (
	HOST      = "localhost"
	TCP_PORT  = "33401"
	INTERFACE = "eth0"
)

func main() {
	emulated := flag.Bool("emu", true,
		"use emulated (true) or real CoordNode (false)")
	flag.Parse()

	dl_chan := make(chan []byte, 1)
	c_ul_chan := make(chan []byte, 1)
	d_ul_chan := make(chan []byte, 1)

	if *emulated {
		go workers.EmulCoordNode(dl_chan, c_ul_chan, d_ul_chan)
	} else {
		// TODO implement serial worker

	}
	// Handle connections in a goroutine
	go listeners.ListenTCP(HOST, TCP_PORT, INTERFACE,
		dl_chan, c_ul_chan, d_ul_chan)
	select {}
}
