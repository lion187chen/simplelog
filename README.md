# Simple Log

A very simple log system(golang).

## Install

```bash
go get github.com/lion187chen/simplelog
```

## Demo

```go
package main

import (
    "time"

    "github.com/lion187chen/simplelog"
)

func CreateLog(file string, LogLevel simplelog.LogLevel) *simplelog.SimpleLog {
    switch file {
    case "":
        return new(simplelog.SimpleLog).InitStd(LogLevel, simplelog.Ltime|simplelog.Lfile|simplelog.Llevel)
    default:
        return new(simplelog.SimpleLog).InitRotating(file, 1024*10, 10, LogLevel)
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

```
