package tcp

import (
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
)

func (t *TCP) Listen() {
	ln, err := net.Listen("tcp", fmt.Sprintf("%s:%d", t.ip, t.portMap.Src))

	if err != nil {
		t.errorChan <- err
		return
	}

	defer ln.Close()

	go func() {
		log.Debugf("[%s] Accepting connections", t.Name())
		for {
			conn, err := ln.Accept()
			log.Debugf("[%s] Accepted connection from: %s", t.Name(), conn.RemoteAddr())

			if err != nil {
				log.Errorf("[%s] Failed accepting connection: %s", t.Name(), err.Error())
				break
			}

			t.connectionChan <- conn
		}
	}()

	<-t.controlChan
}
