package main

import (
	"net"
	"fmt"
	"./device"
	"./tool"
	"./rpc"
	"time"
	"flag"
)

var (
	lginfo     device.LG
	aliveTimer time.Time
	host       string
	port       int
	server     string
)

func init() {
	flag.StringVar(&host, "host", "192.168.1.100", "host ip")
	flag.StringVar(&server, "server", "127.0.0.1", "remote server ip")
	flag.IntVar(&port, "port", 38302, "remote server port")
}

//  time out timer ==300
func timeOutHandler() {

	ticket := time.NewTicker(time.Millisecond * 50)
	for {
		select {
		case <-ticket.C:
			if time.Now().UnixNano()-aliveTimer.UnixNano() > 1e6*320 {
				lginfo.Flag = 0
				lginfo.Ibreak = 153
				fmt.Println("Connect to server side timeout")
				// 300ms no reply then shutdown
			} else {
				//fmt.Println("i am  alive")
			}
		}
	}
}

func writeToVehicle(conn *net.UDPConn) {

	ticket := tool.NewInterval(20);
	for {
		select {
		case <-ticket:
			conn.Write(lginfo.DoSteering())

			if lginfo.Ibreak < 255 {
				conn.Write(lginfo.DoBrake())
			}

			if lginfo.Gas < 255 {
				conn.Write(lginfo.DoThrottle())
			}

			//conn.Write(lginfo.DoGear())
		}
	}
}

/**

lg  can  38300
 can l  38302
*/

func main() {
	flag.Parse()

	/////
	aliveTimer = time.Now()

	lginfo.Init()

	// parse server ip need valid
	serverIp := net.ParseIP(server);
	serverAdd := net.UDPAddr{
		IP:   serverIp,
		Port: 38302,
	}

	// 本机地址
	hostIp := net.ParseIP(host)
	hostAdd := net.UDPAddr{
		IP:   hostIp,
		Port: 38300,
	}
	// Can卡固定IP
	canAddr := net.UDPAddr{
		IP:   net.IPv4(192, 168, 1, 10),
		Port: 8002,
	}
	lg, err := net.DialUDP("udp", nil, &serverAdd)
	tool.CheckError(err)
	can, err := net.DialUDP("udp", &hostAdd, &canAddr)
	tool.CheckError(err)

	defer lg.Close()
	defer can.Close()

	//
	go rpc.KeepAlive(lg)
	go timeOutHandler();
	go writeToVehicle(can)
	//go read(can)

	buf := make([]byte, 12)
	for {
		_, err := lg.Read(buf)
		if err != nil {
			fmt.Print(err)
			continue
		}

		// 读取 状态
		lginfo.ReadDriver(buf)
		aliveTimer = time.Now()
		// 反馈状态
		lg.Write(lginfo.Pong(1))
	}

}

// read can state
func read(conn *net.UDPConn) {

	buf := make([]byte, 13)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			fmt.Print(err)
			continue
		}
		// 读取can 状
		fmt.Println(buf)
	}
}
