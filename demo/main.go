package main

import (
	"github.com/lion187chen/simplelog"
)

var sl *simplelog.SimpleLog

func main() {
	sl = new(simplelog.SimpleLog).InitRotating("./log/MS.log", 1024*1024*2, 100, simplelog.LevelTrace)
}
