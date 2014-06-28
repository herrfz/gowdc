package main

import(
    "flag"
    "github.com/herrfz/gowdc/listeners"
)

const (
	HOST     = "localhost"
	TCP_PORT = "33401"
)

func main() {
    emulated := flag.Bool("emu", true, 
        "use emulated (true) or real CoordNode (false)")
    flag.Parse()

    dl_chan := make(chan []byte, 1)
    ul_chan := make(chan []byte, 1)

    if *emulated {
        go listeners.ListenEmuSerial(dl_chan, ul_chan)
    } else {
        // TODO implement serial worker
        
    }
	// Handle connections in a goroutine
	go listeners.ListenTCP(HOST, TCP_PORT, dl_chan, ul_chan)
	select {}
}
