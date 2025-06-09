# Simple Log

A very simple log system(golang).

## Install

```bash
go get github.com/lion187chen/simplelog
```

## Example

```go
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
```

## 日志头和日志尾

Simple Log 使用关键词使日志文件头尾记录更加明确：

1. StartLog 代表日志系统首次启动，开始记录日志。StartLog 有可能出现在日志文件的中间部分，说明日志被续写到了上次日志文件中。
2. EndLog 代表日志系统关闭。
3. Rotate 代表将要记录日志到下一个文件。
4. StartFile 指 Rotate 产生的新文件。

日志头尾记录非常重要，可用于判断日志文件是否完整，以及 Timed Rotate 日志在文件切换时系统时间是否发生了变化。若无日志头尾记录，由于系统时间变更，可能导致时间上本该连续的 Timed Rotate 日志文件变得不连续。

Simple Log 的 StartLog 可以很好的区分一次日志系统启动，将其和后续的 Rotate 日志区分开。

Timed 日志文件结尾的 Rotate 关键字记录了下一个日志文件的名称，通过这个文件名称可以判断两个时间上看起来不连续的日志（由系统时间变更导致）实际上是否是连续的。

而 File Rotate 日志则是按文件名称中的序号顺序记录的。
