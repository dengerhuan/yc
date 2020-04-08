package device

import (
	"../config"
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
	mu              sync.RWMutex
	seq             int32
	streetingAction *Vehicle //横向
	throttlebAtion  *Vehicle //纵向
	smoothAtion     *Vehicle // 缓动
	gearbuf         [13]byte
	feedback        [6]byte // 0 init  1 res  2keeap  2 other
	temporaryWheel  int32
)

func init() {
	throttlebAtion.Action = "forward"
	throttlebAtion.Value = 0

	streetingAction.Action = "left"
	streetingAction.Value = 0
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
	80 00对应0%   32768
	86 66对应5%   34406
	8C CC 对应10% 36044
	99 99对应20%  39321
	FFFF对应100%  65535
	具体公式 (DATA23-HEX8000)*0.003=油门开度百分比
	*/

	/**
	65536-32768/100  约等于  0x147

	lginfo.gas * 0x147 +0x8000
	*/

	_gas := uint16(255-lginfo.Gas)*0x80 + 0x8000
	if lginfo.Gas < config.MAXTHROTTLE {
		_gas = 0xc000
	}

	// fixed bug bread >0 then throttle =0
	if lginfo.Ibreak < 255 {
		_gas = 0x8000
	}

	slice := throttlebuf[:]

	slice[0] = 0x08

	slice[3] = 0x03
	slice[4] = 0xD6

	slice[6] = 0x1C
	slice[7] = byte(_gas >> 8)
	slice[8] = byte(_gas)

	return slice

}

// 档位没变化是发 0
func (lginfo *LG) DoGear() int8 {

	mu.RLock()
	defer mu.RUnlock()

	//送ID 501 DATA0=01/02/03/04

	// DATA0=01/02/03/04 对应于P1 R2 N3 D4档位

	switch lginfo.IGear {

	// 0 N 3
	// 64 R 2
	// 1 4 16 // 1 3 5 -D 4
	// 2 8 32 // 2 4 6 -P 1
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
