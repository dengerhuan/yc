package main

// version cuc
import (
	"./device"
	"./nrpc"
	"./tool"
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
	flag.StringVar(&endpoint, "endpoint", "tcp://127.0.0.1:5556", "vehicle zmq endpoint")
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

			conn.SendAndReceive([][]byte{})

			// check gear
			// check  flag
			//

			// 检查档位信息， 如果反馈nil 说明 档位没有变化
			// 档位 返回 135  go forward do smooth
			// 档位 返回 N 缓慢 停下
			//  档位 返回 P 档 急停
			//  横向控制 不受档位变化影响

			// 纵向控制 收影响
			// check gear {
			//  go  them smooth
			//  else
			// nil go on
			//}
			//conn.Write(lginfo.DoSteering())
			//
			//if lginfo.Ibreak < 255 {
			//	conn.Write(lginfo.DoBrake())
			//}
			//
			//if lginfo.Gas < 255 {
			//	conn.Write(lginfo.DoThrottle())
			//}

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
	serverIp := net.ParseIP(server)
	serverAdd := net.UDPAddr{
		IP:   serverIp,
		Port: 38302,
	}

	lg, err := net.DialUDP("udp", nil, &serverAdd)
	tool.CheckError(err)
	zmq, err := nrpc.DialZmq("req", endpoint)

	defer lg.Close()

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
