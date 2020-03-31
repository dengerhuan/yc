package rpc

import (
	"net"
	"time"
)

var (
	keepbuf [8]byte
)

func KeepAlive(conn *net.UDPConn) {
	//ticket := tool.NewInterval(50)
	ticket := time.NewTicker(time.Millisecond * 50)

	for {
		select {
		case <-ticket.C:
			_seqbuf := keepbuf[:]
			_seqbuf[4] = 2
			conn.Write(_seqbuf)
		}
	}

}
