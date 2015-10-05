package tcp

import (
	"log"

	"fmt"
	"net"
)

func (l *LoadBalancer) listen(c chan net.Conn, control chan bool, errChan chan error) {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", l.ip, l.port.Src))

	if err != nil {
		errChan <- err
		return
	}

	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()

			if err != nil {
				log.Printf("Failed accepting connection: %s", err.Error())
				break
			}

			c <- conn
		}
	}()

	<-control
}
