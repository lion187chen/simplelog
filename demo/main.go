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

func main() {
	log := CreateLog("./log/MS.log", simplelog.LevelInfo)
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
