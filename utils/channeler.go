package utils

type SocketIn interface {
    Read() ([]byte, error)
}

func MakeChannel(sock SocketIn) <-chan []byte {
    c := make(chan []byte)
    go func() {
        for {
            buf, _ := sock.Read()
            c <- buf
        }
    }()
    return c
}