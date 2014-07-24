// this package implements socket abstractions into channel
package utils

type SocketIn interface {
	Read() ([]byte, error)
}

func MakeChannel(sock SocketIn) <-chan []byte {
	c := make(chan []byte)
	go func() {
		for {
			buf, err := sock.Read()
			if err != nil {
				continue // don't send anything to channel
			} else {
				c <- buf
			}
		}
	}()
	return c
}
