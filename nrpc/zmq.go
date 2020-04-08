package nrpc

import (
	"errors"
	czmq "github.com/zeromq/goczmq"
)

type Zmq struct {
	endpoint string
	sock     *czmq.Sock
}

func DialZmq(zmqType string, endpoint string) (*Zmq, error) {

	switch zmqType {

	case "req":
	default:
		return nil, errors.New("UNSUPPOTTYPE")
	}

	// sock
	reqSock, err := czmq.NewReq(endpoint)
	if err != nil {
		return nil, err
	}
	sock := &Zmq{
		endpoint: endpoint,
		sock:     reqSock,
	}
	return sock, nil
}

func (zmq *Zmq) SendAndReceive(parts [][]byte) error {
	err := zmq.sock.SendMessage(parts)
	if err != nil {
		return err
	}
	_, err = zmq.sock.RecvMessage()
	return err
}

func (zmq *Zmq) SendMessage(parts [][]byte) error {
	return zmq.sock.SendMessage(parts)
}

func (zmq *Zmq) RecvMessage() ([][]byte, error) {
	reply, err := zmq.sock.RecvMessage()
	return reply, err
}

func (zmq *Zmq) Destroy() {
	zmq.sock.Destroy()
}