package main

import (
	"fmt"
	czmq "github.com/zeromq/goczmq"
	"log"

	//"os"
	"strconv"
	"strings"
)

func main() {
	pubEndpoint := "tcp://127.0.0.1:5556"
	totalTemperature := 0

	//filter := "59937"
	//if len(os.Args) > 1 {
	//	filter = string(os.Args[1])
	//}

	//subSock, err := czmq.NewReq(pubEndpoint, filter)
	subSock, err := czmq.NewSub(pubEndpoint, "59937")
	if err != nil {
		panic(err)
	}

	defer subSock.Destroy()

	//fmt.Printf("Collecting updates from weather server for %s…\n", filter)
	subSock.Connect(pubEndpoint)

	for i := 0; i < 100; i++ {

		err := reqSock.SendMessage(msg)
		if err != nil {
			log.Fatalf("Failed to send message: %s", err)
		}

		reply, err := reqSock.RecvMessage()
		if err != nil {
			log.Fatalf("Failed to receive message: %s", err)
		}

		if len(reply) != 1 {
			log.Fatalf("Message of incorrect size received: %d", len(reply))
		}

		fmt.Println(msg)
		if err != nil {
			panic(err)
		}

		weatherData := strings.Split(string(msg), " ")
		temperature, err := strconv.ParseInt(weatherData[1], 10, 64)
		if err == nil {
			totalTemperature += int(temperature)
		}
	}

	fmt.Printf("Average temperature for zipcode %s was %dF \n\n", "59938",
		totalTemperature/100)
}
