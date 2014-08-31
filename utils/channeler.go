// this package implements socket abstractions into channel
package utils

type SocketIn interface {
	Read() ([]byte, error)
	ReadSerial() ([]byte, error)
}

func MakeChannel(sock SocketIn) <-chan []byte {
	c := make(chan []byte)
	go func() {
		for {
			buf, err := sock.Read()
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

func MakeSerialChannel(sock SocketIn) <-chan []byte {
	c := make(chan []byte)
	go func() {
		for {
			buf, err := sock.ReadSerial()
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