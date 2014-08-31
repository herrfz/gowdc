// this package abstracts I/O-device reader into channel
package utils

type DeviceReader interface {
	ReadDevice() ([]byte, error)
}

func MakeChannel(dev DeviceReader) <-chan []byte {
	c := make(chan []byte)
	go func() {
		for {
			buf, err := dev.ReadDevice()
			if err != nil {
				if err.Error() == "DONTPANIC" {
					// non-critical error from caller
					continue
				} else if err.Error() == "PANIC" {
					// critical error
					break
				} else {
					// on all other errors, stop goroutine
					break
				}

			} else {
				c <- buf
			}
		}
	}()
	return c
}
