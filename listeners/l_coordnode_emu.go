// This package implements WDC comm listeners
package listeners

func ListenEmuSerial(dl_chan, ul_chan chan []byte) {
    var n int
    ret_msg := make([]byte, 256)

    for {
        buf := <- dl_chan
        // TODO: do the work
        switch buf[1] {
        case 0x01:
            n = 10
            ret_msg[0] = 9
            ret_msg[1] = 2
            ret_msg[2] = 3

        case 0x03:
            n = 2
            ret_msg[0] = 1
            ret_msg[1] = 4

        case 0x07:
            n = 2
            ret_msg[0] = 1
            ret_msg[1] = 8

        case 0x11:
            n = 1
        }
        ul_chan <- ret_msg[:n]
    }
}