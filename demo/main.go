package main

import (
	"time"

	"github.com/lion187chen/simplelog"
)

var sl *simplelog.SimpleLog

func main() {
	sl = new(simplelog.SimpleLog).InitRotating("./log/MS.log", 1024*10, 10, simplelog.LevelTrace)
	for i := 0; i < 10000000; i++ {
		sl.Debug("hello world")
		time.Sleep(8 * time.Millisecond)
	}
}
