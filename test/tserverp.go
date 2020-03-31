package main

import (
	"fmt"
	"log"
	"math/rand"

	czmq "github.com/zeromq/goczmq"
)

//https://github.com/zeromq/goczmq
/**


需要安装配置
libsodium
libzmq
czmq
*/

func main() {
	endpoint := "tcp://127.0.0.1:5556"
	repSock, err := czmq.NewRep(endpoint)
	if err != nil {
		panic(err)
	}

	defer repSock.Destroy()

	for {

		msg, err := repSock.RecvMessage()
		if err != nil {
			log.Fatalf("Failed to receive message: %s", err)
		}

		if len(msg) != 1 {
			log.Fatalf("Message of incorrect size received: %d", len(msg))
		}
		zipcode := rand.Intn(100000)
		temperature := rand.Intn(215) - 85
		relHumidity := rand.Intn(50) + 10

		rmsg := fmt.Sprintf("%d %d %d", zipcode, temperature, relHumidity)
		err = repSock.SendMessage([][]byte{[]byte(rmsg)})
		if err != nil {
			panic(err)
		}
	}
}

////**
//
//// Send messages and read replies.
//for i := 0; i != *roundtripCount; i++ {
//err := reqSock.SendMessage(msg)
//if err != nil {
//log.Fatalf("Failed to send message: %s", err)
//}
//
//reply, err := reqSock.RecvMessage()
//if err != nil {
//log.Fatalf("Failed to receive message: %s", err)
//}
//
//if len(reply) != 1 {
//log.Fatalf("Message of incorrect size received: %d", len(reply))
//}
//
//if len(reply[0]) != *messageSize {
//log.Fatalf("Message of incorrect size received: %d", len(reply[0]))
//}
//}
