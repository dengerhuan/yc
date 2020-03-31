package tool

import (
	"time"
)

// not complete
type Interviel struct {
	Channel chan int64
	flag    bool
}

func NewInterval(duration int64) chan int64 {

	last := time.Now().UnixNano();
	var i int64 = 0
	ch := make(chan int64)

	_duration := duration * 1e6

	go func() {
		for {
			if time.Now().UnixNano()-last > _duration {

				last = time.Now().UnixNano()

				//fmt.Println(time.Now().UnixNano())
				ch <- i
				i++
			} else {
				time.Sleep(time.Microsecond * 100)
			}
		}
	}()
	return ch
}

func CloseInterval(ch chan int64) {
	close(ch)
}
