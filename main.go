package main

// version cuc
import (
	"./device"
	"./nrpc"
	"./tool"
	"encoding/json"
	"flag"
	"fmt"
	"math"
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
	current_steering device.Vehicle
	last_steering_value int16
	change_steering_flag bool
)

func init() {

	flag.StringVar(&host, "host", "192.168.56.101", "host ip")
	flag.StringVar(&server, "server", "192.168.56.1", "remote server ip")
	flag.IntVar(&port, "port", 38302, "remote server port")
	flag.StringVar(&endpoint, "endpoint", "tcp://127.0.0.1:5555", "vehicle zmq endpoint")
	last_steering_value = 0
	change_steering_flag = false
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
	if err == nil{
		fmt.Println("UDP server connection successful")
	}

	zmq, err := nrpc.DialZmq("req", endpoint)
	tool.CheckError(err)
	if err == nil{
		fmt.Println("ZMQ connection successful")
	}

	defer lg.Close()
	defer zmq.Destroy()

	setSmooth(zmq)
	//
	go nrpc.KeepAlive(lg)
	go timeOutHandler()
	go writeToVehicle(zmq)

	buf := make([]byte, 12)
	for {
		//fmt.Printf("received buffer : %x\n",buf)
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
				err := json.Unmarshal(steering,&current_steering)
				if err != nil{
					tool.CheckError(err)
				}
				if math.Abs(float64(current_steering.Value - last_steering_value)) > 100{
					last_steering_value = current_steering.Value
					fmt.Println(string(steering))
					conn.SendAndReceiveFrame(steering)
					change_steering_flag = true
				}else{
					change_steering_flag = false
				}

			}

			// 油门
			throttle := lginfo.DoThrottle()
			if throttle != nil {
				fmt.Println(string(throttle))
				if change_steering_flag == false{
					conn.SendAndReceiveFrame(throttle)
				}

			}

			// 刹车
			ibreak := lginfo.DoBrake()
			if ibreak != nil {
				fmt.Println(string(ibreak))
				conn.SendAndReceiveFrame(ibreak)
			}

		}
	}
}
