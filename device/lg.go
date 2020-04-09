package device

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"log"
	"math"
	"sync"
)

type LG struct {
	Wheel  int32 //0-65535
	Gas    byte  //255 -0
	Ibreak byte  // 255- 0
	IGear  byte  //
	Flag   byte  // 1 control 0 out control
}

var (
	mu  sync.RWMutex
	seq int32

	streetingAction *Vehicle //横向
	throttlebAtion  *Vehicle //纵向
	SmoothAction    *Vehicle // 缓动
	stopAction      *Vehicle //停车

	feedback       [6]byte // 0 init  1 res  2keeap  2 other
	temporaryWheel int32
)

func init() {
	throttlebAtion = &Vehicle{Action: "forward"}
	streetingAction = &Vehicle{Action: "right"}
	stopAction = &Vehicle{Action: "stop"}
	SmoothAction = &Vehicle{Action: "smooth", Value: 100}
}

// DATA0=01/02/03/04 对应于PRND档位
func (lginfo *LG) Init() {
	lginfo.Gas = 0
	lginfo.Ibreak = 0
	lginfo.IGear = 1
	lginfo.Wheel = 32768
}

/**
0  -54
*/
func (lginfo *LG) Pong(state byte) []byte {
	_seqbuf := feedback[:]
	_seqbuf[0] = byte(seq)
	_seqbuf[1] = byte(seq >> 8)
	_seqbuf[2] = byte(seq >> 16)
	_seqbuf[3] = byte(seq >> 24)
	_seqbuf[4] = state
	return _seqbuf
}

func (lginfo *LG) ReadDriver(buf []byte) {
	read := bytes.NewReader(buf)
	mu.Lock()
	defer mu.Unlock()
	//
	binary.Read(read, binary.LittleEndian, &seq)
	binary.Read(read, binary.LittleEndian, &temporaryWheel)
	binary.Read(read, binary.LittleEndian, &lginfo.Gas)
	binary.Read(read, binary.LittleEndian, &lginfo.Ibreak)
	binary.Read(read, binary.LittleEndian, &lginfo.IGear)
	binary.Read(read, binary.LittleEndian, &lginfo.Flag)
	lginfo.holerSteering()
	//fmt.Println(lginfo)
}

func (lginfo *LG) holerSteering() {

	if math.Abs(float64(temporaryWheel-lginfo.Wheel)) < 300 {
		lginfo.Wheel = temporaryWheel

	} else if temporaryWheel-lginfo.Wheel > 300 {

		lginfo.Wheel += 300
	} else {
		lginfo.Wheel -= 300
	}
}

// 横向
func (lginfo *LG) DoSteering() []byte {
	mu.RLock()
	defer mu.RUnlock()

	// 转换 lg info to +-1000
	_w := float32(lginfo.Wheel)/32.7675 - 1000
	_wheel := int16(-_w)
	// fixed bug limit wheel range
	if _wheel < -1000 {
		_wheel = -1000
	} else if _wheel > 1000 {
		_wheel = 1000
	}

	streetingAction.Value = _wheel
	action, err := json.Marshal(streetingAction)

	if err != nil {
		log.Fatal(err)
		return nil
	}
	return action
}

// 纵向
func (lginfo *LG) DoThrottle() []byte {

	mu.RLock()
	defer mu.RUnlock()

	/**
	  255 对应0
	*/

	acceleration := int16((255 - lginfo.Gas) * 4)

	if acceleration > 1000 {
		acceleration = 1000
	}

	switch lginfo.DoGear() {
	// r
	case 2:
		acceleration = acceleration * -1
		// d
	case 4:

	default:
		acceleration = 0
	}

	throttlebAtion.Value = acceleration
	action, err := json.Marshal(throttlebAtion)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return action
}

// 0 缓动  1 急停 2 自由停
/**
刹车 没有数据时 返回 nil
刹车到4/5 时 极停
刹车到2/5 时 缓动
轻踩刹车时 自由停
*/
func (lginfo *LG) DoBrake() []byte {

	mu.RLock()
	defer mu.RUnlock()
	iBbreak := 255 - lginfo.Ibreak

	var stopMode int16
	switch {
	case iBbreak == 0:
		return nil
	case iBbreak > 200:

		stopMode = 1
	case iBbreak > 100:
		stopMode = 0
	default:
		stopMode = 2
	}

	stopAction.Value = stopMode

	action, err := json.Marshal(stopAction)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return action
}

// 档位没变化是发 0
func (lginfo *LG) DoGear() int8 {

	mu.RLock()
	defer mu.RUnlock()

	//送ID 501 DATA0=01/02/03/04
	// DATA0=01/02/03/04 对应于P1 R2 N3 D4档位
	// 0 N 3
	// 64 R 2
	// 1 4 16 // 1 3 5 -D 4
	// 2 8 32 // 2 4 6 -P 1

	switch lginfo.IGear {
	case 0:
		return 3
	case 64:
		return 2
	case 1, 4, 16:
		return 4
	case 2, 8, 32:
		return 1
	default:
		return 0
	}
}
