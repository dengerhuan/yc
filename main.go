package main

// version cuc
import (
	"./device"
	"./nrpc"
	"./tool"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"time"
)

var (
	lginfo     device.LG
	aliveTimer time.Time
	host       string
	port       int
	server     string
	endpoint   string
)

func init() {

	flag.StringVar(&host, "host", "192.168.1.100", "host ip")
	flag.StringVar(&server, "server", "127.0.0.1", "remote server ip")
	flag.IntVar(&port, "port", 38302, "remote server port")
	flag.StringVar(&endpoint, "endpoint", "tcp://127.0.0.1:5555", "vehicle zmq endpoint")
}

/**

lg  can  38300
 can l  38302
*/

func main() {
	flag.Parse()
	aliveTimer = time.Now()
	lginfo.Init()
	// parse server ip need valid
	serverIp := net.ParseIP(server)
	serverAdd := net.UDPAddr{
		IP:   serverIp,
		Port: 38302,
	}

	lg, err := net.DialUDP("udp", nil, &serverAdd)
	tool.CheckError(err)

	zmq, err := nrpc.DialZmq("req", endpoint)
	tool.CheckError(err)

	defer lg.Close()
	defer zmq.Destroy()

	setSmooth(zmq)
	//
	go nrpc.KeepAlive(lg)
	go timeOutHandler()
	go writeToVehicle(zmq)

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

/**
当前缓动仅进行一次设置，可以尝试与其他动作一起发送到 vehicle 参考 writeToVehicle
*/
func setSmooth(conn *nrpc.Zmq) {
	data, _ := json.Marshal(device.SmoothAction)
	err := conn.SendAndReceiveFrame(data)
	tool.CheckError(err)
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

// customer change
func writeToVehicle(conn *nrpc.Zmq) {

	//
	ticket := tool.NewInterval(50)
	for {
		select {
		case <-ticket:

			// 横向
			steering := lginfo.DoSteering()

			if steering != nil {
				conn.SendAndReceiveFrame(steering)
			}

			// 油门
			throttle := lginfo.DoThrottle()
			if throttle != nil {
				conn.SendAndReceiveFrame(throttle)
			}

			// 刹车
			ibreak := lginfo.DoBrake()
			if ibreak != nil {
				conn.SendAndReceiveFrame(ibreak)
			}

		}
	}
}
