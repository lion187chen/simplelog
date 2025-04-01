// Expanded from github.com/siddontang/go/tree/master/log
//
// It also supports different log handlers which you can log to stdout, file, socket, etc...
//
// Usage
//
//	import "blacktea.vip.cpolar.top/OrgGo/simplelog"
//
//	//log with different level
//	simplelog.Info("hello world")
//	simplelog.Error("hello world")
//
//	//create a simplelog with specified handler
//	h := NewStreamHandle(os.Stdout)
//	l := simplelog.NewDefault(h)
//	l.Info("hello world")
//	l.Infof("%s %d", "hello", 123)
package simplelog

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type LogLevel int

// log level, from low to high, more higher means more serious
const (
	LevelTrace LogLevel = iota
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

const (
	Ltime  = 1 << iota //print timestamp format as "2006/01/02 15:04:05.000"
	Lfile              // print file name and line format as file.go:123
	Llevel             // print log level format as [Trace|Debug|Info...]
)

var LevelName [6]string = [6]string{"Trace", "Debug", "Info ", "Warn ", "Error", "Fatal"}

const TimeFormat = "2006/01/02 15:04:05.000"

const maxBufPoolSize = 16

type atomicInt32 int32

func (i *atomicInt32) Set(n int) {
	atomic.StoreInt32((*int32)(i), int32(n))
}

func (i *atomicInt32) Get() int {
	return int(atomic.LoadInt32((*int32)(i)))
}

type SimpleLog struct {
	level atomicInt32
	flag  int

	hMutex  sync.Mutex
	handler StreamHandler

	bufMutex sync.Mutex
	bufs     [][]byte

	closed atomicInt32
}

// 初始化函数
// 参数 name 为 log 文件名
// 参数 level 为：
//
//	LevelTrace
//	LevelDebug
//	LevelInfo
//	LevelWarn
//	LevelError
//	LevelFatal
//	其中之一
//
// 参数 flag 可以为
//
//	log.Ltime
//	log.Lfile
//	log.Llevel
//	的组合
func (l *SimpleLog) Init(handler StreamHandler, level LogLevel, flag int) *SimpleLog {
	l.level.Set(int(level))
	l.handler = handler

	l.flag = flag

	l.closed.Set(0)

	l.bufs = make([][]byte, 0, 16)
	return l
}

func (l *SimpleLog) InitStd(level LogLevel, flag int) *SimpleLog {
	handler, e := NewStreamHandle(os.Stdout)
	if e != nil {
		panic(e)
	}

	return l.Init(handler, level, flag)
}

func (l *SimpleLog) InitFile(name string, level LogLevel, flag int) *SimpleLog {
	handler, e := new(FileHandler).InitFile(name)
	if e != nil {
		panic(e)
	}

	return l.Init(handler, level, flag)
}

func (l *SimpleLog) InitRotating(name string, maxBytes, backupCount int, level LogLevel) *SimpleLog {
	handler, e := new(RotatingFileHandler).InitRotating(name, maxBytes, backupCount)
	if e != nil {
		panic(e)
	}

	return l.Init(handler, level, Ltime|Lfile|Llevel)
}

func (l *SimpleLog) InitTimedRotating(name string, when int8, interval int, level LogLevel) *SimpleLog {
	handler, e := new(TimedRotatingFileHandler).InitTimedRotating(name, when, interval)
	if e != nil {
		panic(e)
	}

	return l.Init(handler, level, Ltime|Lfile|Llevel)
}

func (l *SimpleLog) popBuf() []byte {
	l.bufMutex.Lock()
	var buf []byte
	if len(l.bufs) == 0 {
		buf = make([]byte, 0, 1024)
	} else {
		buf = l.bufs[len(l.bufs)-1]
		l.bufs = l.bufs[0 : len(l.bufs)-1]
	}
	l.bufMutex.Unlock()

	return buf
}

func (l *SimpleLog) putBuf(buf []byte) {
	l.bufMutex.Lock()
	if len(l.bufs) < maxBufPoolSize {
		buf = buf[0:0]
		l.bufs = append(l.bufs, buf)
	}
	l.bufMutex.Unlock()
}

func (l *SimpleLog) Close() {
	if l.closed.Get() == 1 {
		return
	}
	l.closed.Set(1)

	l.handler.Close()
}

// set log level, any log level less than it will not log
func (l *SimpleLog) SetLevel(level LogLevel) {
	l.level.Set(int(level))
}

// name can be in ["trace", "debug", "info", "warn", "error", "fatal"]
func (l *SimpleLog) SetLevelByName(name string) {
	name = strings.ToLower(name)
	switch name {
	case "trace":
		l.SetLevel(LevelTrace)
	case "debug":
		l.SetLevel(LevelDebug)
	case "info":
		l.SetLevel(LevelInfo)
	case "warn":
		l.SetLevel(LevelWarn)
	case "error":
		l.SetLevel(LevelError)
	case "fatal":
		l.SetLevel(LevelFatal)
	}
}

func (l *SimpleLog) SetHandler(h StreamHandler) {
	if l.closed.Get() == 1 {
		return
	}

	l.hMutex.Lock()
	if l.handler != nil {
		l.handler.Close()
	}
	l.handler = h
	l.hMutex.Unlock()
}

func (l *SimpleLog) Output(callDepth int, level LogLevel, format string, v ...interface{}) {
	if l.closed.Get() == 1 {
		// closed
		return
	}

	if l.level.Get() > int(level) {
		// higher level can be logged
		return
	}

	var s string
	if format == "" {
		s = fmt.Sprint(v...)
	} else {
		s = fmt.Sprintf(format, v...)
	}

	buf := l.popBuf()

	if (l.flag&Llevel > 0) || (l.flag&Ltime > 0) || (l.flag&Lfile > 0) {
		buf = append(buf, '[')
	}

	if l.flag&Llevel > 0 {
		buf = append(buf, LevelName[level]...)
	}

	if (l.flag&Llevel > 0) && (l.flag&Ltime > 0) {
		buf = append(buf, " | "...)
	}

	if l.flag&Ltime > 0 {
		now := time.Now().Format(TimeFormat)
		buf = append(buf, now...)
	}

	if (l.flag&Ltime > 0) && (l.flag&Lfile > 0) {
		buf = append(buf, " | "...)
	}

	if l.flag&Lfile > 0 {
		_, file, line, ok := runtime.Caller(callDepth)
		if !ok {
			file = "???"
			line = 0
		} else {
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					file = file[i+1:]
					break
				}
			}
		}

		buf = append(buf, file...)
		buf = append(buf, ':')

		buf = strconv.AppendInt(buf, int64(line), 10)

		if (l.flag&Llevel > 0) || (l.flag&Ltime > 0) || (l.flag&Lfile > 0) {
			buf = append(buf, "] "...)
		}
	}

	buf = append(buf, s...)

	if s[len(s)-1] != '\n' {
		buf = append(buf, '\n')
	}

	// l.msg <- buf

	l.hMutex.Lock()
	l.handler.Write(buf)
	l.hMutex.Unlock()
	l.putBuf(buf)
}

// log with Trace level
func (l *SimpleLog) Trace(v ...interface{}) {
	l.Output(2, LevelTrace, "", v...)
}

// log with Debug level
func (l *SimpleLog) Debug(v ...interface{}) {
	l.Output(2, LevelDebug, "", v...)
}

// log with info level
func (l *SimpleLog) Info(v ...interface{}) {
	l.Output(2, LevelInfo, "", v...)
}

// log with warn level
func (l *SimpleLog) Warn(v ...interface{}) {
	l.Output(2, LevelWarn, "", v...)
}

// log with error level
func (l *SimpleLog) Error(v ...interface{}) {
	l.Output(2, LevelError, "", v...)
}

// log with fatal level
func (l *SimpleLog) Fatal(v ...interface{}) {
	l.Output(2, LevelFatal, "", v...)
}

// log with Trace level
func (l *SimpleLog) Tracef(format string, v ...interface{}) {
	l.Output(2, LevelTrace, format, v...)
}

// log with Debug level
func (l *SimpleLog) Debugf(format string, v ...interface{}) {
	l.Output(2, LevelDebug, format, v...)
}

// log with info level
func (l *SimpleLog) Infof(format string, v ...interface{}) {
	l.Output(2, LevelInfo, format, v...)
}

// log with warn level
func (l *SimpleLog) Warnf(format string, v ...interface{}) {
	l.Output(2, LevelWarn, format, v...)
}

// log with error level
func (l *SimpleLog) Errorf(format string, v ...interface{}) {
	l.Output(2, LevelError, format, v...)
}

// log with fatal level
func (l *SimpleLog) Fatalf(format string, v ...interface{}) {
	l.Output(2, LevelFatal, format, v...)
}
