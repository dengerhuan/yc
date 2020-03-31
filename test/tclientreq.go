package main

import (
	"fmt"
	czmq "github.com/zeromq/goczmq"
	"log"
	"strconv"
	"strings"
)

func main() {
	endpoint := "tcp://127.0.0.1:5556"
	totalTemperature := 0

	reqSock, err := czmq.NewReq(endpoint)
	if err != nil {
		panic(err)
	}
	defer reqSock.Destroy()

	msg := [][]byte{make([]byte, 10)}
	for i := 0; i < 100; i++ {

		err := reqSock.SendMessage(msg)
		if err != nil {
			log.Fatalf("Failed to send message: %s", err)
		}

		reply, err := reqSock.RecvMessage()

		fmt.Println(reply[0])
		if err != nil {
			log.Fatalf("Failed to receive message: %s", err)
		}

		if len(reply) != 1 {
			log.Fatalf("Message of incorrect size received: %d", len(reply))
		}
		weatherData := strings.Split(string(reply[0]), " ")
		temperature, err := strconv.ParseInt(weatherData[1], 10, 64)
		if err == nil {
			totalTemperature += int(temperature)
		}
	}

	fmt.Printf("Average temperature for zipcode %s was %dF \n\n", "59938",
		totalTemperature/100)
}
