package main

import (
	"time"

	"github.com/lion187chen/simplelog"
)

func CreateLog(file string, level simplelog.Level) *simplelog.Log {
	switch file {
	case "":
		return new(simplelog.Log).InitStd(level, simplelog.Ltime|simplelog.Lfile|simplelog.Llevel)
	default:
		return new(simplelog.Log).InitRotating(file, 1024*10, 10, level)
	}
}

func test_timed(name string) {
	log := new(simplelog.Log).InitTimed(name, simplelog.WhenMinute, 2, simplelog.LevelDebug)
	n := time.Now()
	for time.Since(n) < 10*time.Minute {
		log.Debug("hello world")
		time.Sleep(30 * time.Second)
	}
}

func test_timedRotating(name string) {
	log := new(simplelog.Log).InitTimedRotating(name, simplelog.WhenMinute, 2, 3, simplelog.LevelDebug)
	n := time.Now()
	for time.Since(n) < 10*time.Minute {
		log.Debug("hello world")
		time.Sleep(30 * time.Second)
	}
}

func test_rotating(name string) {
	log := CreateLog(name, simplelog.LevelInfo)
	for i := 0; i < 10000000; i++ {
		log.Trace("hello world")
		log.Debug("hello world")
		log.Info("hello world")
		log.Warn("hello world")
		log.Error("hello world")
		log.Fatal("hello world")
		time.Sleep(8 * time.Millisecond)
	}
}

func main() {
	go test_rotating("./log/demo.log")
	go test_timedRotating("./trlog/demo.log")
	test_timed("./tlog/demo.log")
}
