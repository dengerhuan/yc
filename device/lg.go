package device

import (
	"../config"
	"bytes"
	"encoding/binary"
	"math"
	"sync"
)

type LG struct {
	Wheel  int32
	Gas    byte
	Ibreak byte
	IGear  byte
	Flag   byte // 1 control 0 out control
}

var (
	mu           sync.RWMutex
	seq          int32
	streetingbuf [13]byte
	throttlebuf  [13]byte
	ibrakebuf    [13]byte
	gearbuf      [13]byte
	feedback     [6]byte // 0 init  1 res  2keeap  2 other

	temporaryWheel int32;
)

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
	read := bytes.NewReader(buf);
	mu.Lock()
	defer mu.Unlock()
	//
	binary.Read(read, binary.LittleEndian, &seq);
	binary.Read(read, binary.LittleEndian, &temporaryWheel);
	binary.Read(read, binary.LittleEndian, &lginfo.Gas);
	binary.Read(read, binary.LittleEndian, &lginfo.Ibreak);
	binary.Read(read, binary.LittleEndian, &lginfo.IGear);
	binary.Read(read, binary.LittleEndian, &lginfo.Flag);
	lginfo.holerSteering();
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

func (lginfo *LG) DoSteering() []byte {
	mu.RLock()
	defer mu.RUnlock()
	// 转换 lg info to +-540
	_w := float32(lginfo.Wheel)/6.068 - 5400
	_wheel := int16(-_w)
	// fixed bug limit wheel range
	if _wheel < -5400 {
		_wheel = -5400
	} else if _wheel > 5400 {
		_wheel = 5400
	}

	slice := streetingbuf[:]

	slice[0] = 0x08
	slice[3] = 0x01
	slice[4] = 0x12

	slice[5] = lginfo.Flag
	slice[6] = byte(_wheel >> 8)
	slice[7] = byte(_wheel)
	return slice
}

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

	_gas := uint16(255-lginfo.Gas)*0x80 + 0x8000;
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
func (lginfo *LG) DoBrake() []byte {

	mu.RLock()
	defer mu.RUnlock()

	/**
	8C 00对应0m/ss; // 140
	82 00对应 0.5m/ss;
	78 00对应1m/ss;
	6E 00对应1.5m/ss;
	64 00对应2m/ss
	28 00对应5m/ss  40
	 */
	slice := ibrakebuf[:]

	slice[0] = 0x08

	slice[3] = 0x02
	slice[4] = 0xBF
	slice[6] = 0x82
	slice[7] = 0x80
	slice[8] = 0x12
	slice[10] = 0x01

	//ss := lginfo.Ibreak / 255 * 100
	ibreak := float32(lginfo.Ibreak)/255*100 + 0x28;


	var gea byte = byte(ibreak)

	slice[11] = gea



	return slice

}

func (lginfo *LG) DoGear() []byte {

	mu.RLock()
	defer mu.RUnlock()
	//送ID 501 DATA0=01/02/03/04

	// DATA0=01/02/03/04 对应于PRND档位
	slice := gearbuf[:]
	slice[0] = 0x08
	slice[3] = 0x05
	slice[4] = 0x01

	slice[5] = lginfo.Ibreak

	return slice
}
