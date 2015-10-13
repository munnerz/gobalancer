package tcp

import (
	"log"

	"fmt"
	"net"
)

func (t *TCP) Listen() {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", t.ip, t.port))

	if err != nil {
		t.errorChan <- err
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

			t.connectionChan <- conn
		}
	}()

	<-t.connectionChan
}
