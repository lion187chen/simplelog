package main

import (
	"time"

	"github.com/lion187chen/simplelog"
)

func test_timed(name string) {
	log := new(simplelog.Log).InitTimed(name, simplelog.WhenMinute, 2, simplelog.LevelDebug)
	n := time.Now()
	for time.Since(n) < 10*time.Minute {
		log.Trace("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Debug("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Info("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Warn("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Error("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Fatal("hello world")
		time.Sleep(100 * time.Millisecond)
	}
	log.Close()
}

func test_timedRotating(name string) {
	log := new(simplelog.Log).InitTimedRotating(name, simplelog.WhenMinute, 2, 3, simplelog.LevelDebug)
	n := time.Now()
	for time.Since(n) < 10*time.Minute {
		log.Trace("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Debug("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Info("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Warn("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Error("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Fatal("hello world")
		time.Sleep(100 * time.Millisecond)
	}
	log.Close()
}

func test_rotating(name string) {
	log := new(simplelog.Log).InitRotating(name, 1024*10, 10, simplelog.LevelInfo)
	for i := 0; i < 10000000; i++ {
		log.Trace("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Debug("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Info("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Warn("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Error("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Fatal("hello world")
		time.Sleep(100 * time.Millisecond)
	}
	log.Close()
}

func test_file(name string) {
	log := new(simplelog.Log).InitFile(name, simplelog.LevelDebug)
	n := time.Now()
	for time.Since(n) < 10*time.Minute {
		log.Trace("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Debug("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Info("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Warn("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Error("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Fatal("hello world")
		time.Sleep(100 * time.Millisecond)
	}
	log.Close()
}

func test_std() {
	log := new(simplelog.Log).InitStd(simplelog.LevelDebug, simplelog.Ltime|simplelog.Lfile|simplelog.Llevel)
	n := time.Now()
	for time.Since(n) < 10*time.Minute {
		log.Trace("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Debug("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Info("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Warn("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Error("hello world")
		time.Sleep(100 * time.Millisecond)
		log.Fatal("hello world")
		time.Sleep(100 * time.Millisecond)
	}
	log.Close()
}

func main() {
	go test_std()
	go test_file("./log/demo.log")
	go test_rotating("./rlog/demo.log")
	go test_timedRotating("./trlog/demo.log")
	test_timed("./tlog/demo.log")
}
