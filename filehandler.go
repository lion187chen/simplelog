package simplelog

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func isFileExist(filename string) bool {
	info, e := os.Stat(filename)

	if e == nil {
		return !info.IsDir()
	} else {
		return os.IsExist(e)
	}
}

// FileHandler writes log to a file.
type FileHandler struct {
	fd *os.File
}

func (h *FileHandler) InitFile(name string) (*FileHandler, error) {
	dir := path.Dir(name)
	os.Mkdir(dir, 0777)

	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	h.fd = f

	return h, nil
}

func (h *FileHandler) Write(b []byte) (n int, err error) {
	return h.fd.Write(b)
}

func (h *FileHandler) Close() error {
	return h.fd.Close()
}

// RotatingFileHandler writes log a file, if file size exceeds maxBytes,
// it will backup current file and open a new one.
//
// max backup file number is set by backupCount, it will delete oldest if backups too many.
type RotatingFileHandler struct {
	fd *os.File

	fileName    string
	maxBytes    int64
	curBytes    int64
	backupCount uint
}

func (h *RotatingFileHandler) InitRotating(name string, maxBytes int64, backupCount uint) (*RotatingFileHandler, error) {
	h.fileName = name
	h.maxBytes = maxBytes
	h.backupCount = backupCount

	if isFileExist(name) {
		var err error
		h.fd, err = os.OpenFile(name, os.O_CREATE|os.O_RDONLY, 0666)
		if err != nil {
			return nil, err
		}
		f, err := h.fd.Stat()
		if err != nil {
			h.fd.Close()
			return nil, err
		}
		if f.Size() < h.maxBytes {
			h.fd.Seek(0, io.SeekEnd)
		} else {
			h._doRollover()
		}
		return h, nil
	}

	dir := path.Dir(name)
	os.MkdirAll(dir, 0777)

	if maxBytes <= 0 {
		return nil, fmt.Errorf("invalid max bytes")
	}

	var err error
	h.fd, err = os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	f, err := h.fd.Stat()
	if err != nil {
		return nil, err
	}
	h.curBytes = f.Size()

	return h, nil
}

func (h *RotatingFileHandler) Write(p []byte) (n int, err error) {
	h.doRollover()
	n, err = h.fd.Write(p)
	h.curBytes += int64(n)
	return
}

func (h *RotatingFileHandler) Close() error {
	if h.fd != nil {
		return h.fd.Close()
	}
	return nil
}

func (h *RotatingFileHandler) doRollover() {

	if h.curBytes < h.maxBytes {
		return
	}

	f, err := h.fd.Stat()
	if err != nil {
		return
	}

	if h.maxBytes <= 0 {
		return
	} else if f.Size() < int64(h.maxBytes) {
		h.curBytes = f.Size()
		return
	}

	h._doRollover()
}

func (h *RotatingFileHandler) _doRollover() {
	if h.backupCount > 0 {
		h.fd.Close()

		for i := h.backupCount - 1; i > 0; i-- {
			sfn := fmt.Sprintf("%s.%d", h.fileName, i)
			dfn := fmt.Sprintf("%s.%d", h.fileName, i+1)

			os.Rename(sfn, dfn)
		}

		dfn := fmt.Sprintf("%s.1", h.fileName)
		os.Rename(h.fileName, dfn)

		h.fd, _ = os.OpenFile(h.fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		h.curBytes = 0
	} else {
		h.fd.Close()
		h.fd, _ = os.OpenFile(h.fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
		h.curBytes = 0
	}
}

// TimedFileHandler writes log to a file.
type TimedFileHandler struct {
	fd *os.File

	dir        string
	name       string
	interval   time.Duration
	suffix     string
	rolloverAt time.Time
}

type WhenInterval int

const (
	WhenSecond WhenInterval = iota
	WhenMinute
	WhenHour
	WhenDay
)

func (h *TimedFileHandler) InitTimed(name string, when WhenInterval, interval int64) (*TimedFileHandler, error) {
	dir := path.Dir(name)
	os.Mkdir(dir, 0777)

	h.dir, h.name = filepath.Split(name)

	switch when {
	case WhenSecond:
		h.interval = 1 * time.Second
		h.suffix = "2006-01-02-15-04-05"
	case WhenMinute:
		h.interval = 1 * time.Minute
		h.suffix = "2006-01-02-15-04"
	case WhenHour:
		h.interval = 1 * time.Hour
		h.suffix = "2006-01-02-15"
	case WhenDay:
		h.interval = 1 * time.Hour * 24
		h.suffix = "2006-01-02"
	default:
		return nil, fmt.Errorf("invalid when_rotate: %d", when)
	}

	h.interval = time.Duration(interval) * h.interval

	h.rolloverAt = time.Now()
	var err error
	h.fd, err = os.OpenFile(filepath.Join(h.dir, h.rolloverAt.Format(h.suffix)+"_"+h.name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *TimedFileHandler) doRollover() {
	if time.Since(h.rolloverAt) > time.Duration(h.interval) {
		h.fd.Close()
		h.rolloverAt = time.Now()
		h.fd, _ = os.OpenFile(filepath.Join(h.dir, h.rolloverAt.Format(h.suffix)+"_"+h.name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
}

func (h *TimedFileHandler) Write(b []byte) (n int, err error) {
	h.doRollover()
	return h.fd.Write(b)
}

func (h *TimedFileHandler) Close() error {
	return h.fd.Close()
}

type TimedRotatingFileHandler struct {
	fd *os.File

	dir         string
	name        string
	interval    time.Duration
	suffix      string
	rolloverAt  time.Time
	backupCount uint
}

func (h *TimedRotatingFileHandler) InitTimedRotating(name string, when WhenInterval, interval int64, backupCount uint) (*TimedRotatingFileHandler, error) {
	h.backupCount = backupCount
	dir := path.Dir(name)
	os.Mkdir(dir, 0777)

	h.dir, h.name = filepath.Split(name)

	switch when {
	case WhenSecond:
		h.interval = 1 * time.Second
		h.suffix = "2006-01-02-15-04-05"
	case WhenMinute:
		h.interval = 1 * time.Minute
		h.suffix = "2006-01-02-15-04"
	case WhenHour:
		h.interval = 1 * time.Hour
		h.suffix = "2006-01-02-15"
	case WhenDay:
		h.interval = 1 * time.Hour * 24
		h.suffix = "2006-01-02"
	default:
		return nil, fmt.Errorf("invalid when_rotate: %d", when)
	}

	h.interval = time.Duration(interval) * h.interval

	h.rolloverAt = time.Now()
	var err error
	h.fd, err = os.OpenFile(filepath.Join(h.dir, h.rolloverAt.Format(h.suffix)+"_"+h.name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	h._doRollover()
	return h, nil
}

func (h *TimedRotatingFileHandler) _doRollover() {
	fs, err := h.ListDir(h.dir, h.suffix)
	if err != nil {
		return
	}
	if len(fs) > int(h.backupCount) {
		for i := 0; i < len(fs)-int(h.backupCount); i++ {
			os.Remove(fs[i])
		}
	}
}

func (h *TimedRotatingFileHandler) doRollover() {
	if time.Since(h.rolloverAt) > time.Duration(h.interval) {
		h.fd.Close()
		h.rolloverAt = time.Now()
		h.fd, _ = os.OpenFile(filepath.Join(h.dir, h.rolloverAt.Format(h.suffix)+"_"+h.name), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
	h._doRollover()
}

func (h *TimedRotatingFileHandler) Write(b []byte) (n int, err error) {
	h.doRollover()
	return h.fd.Write(b)
}

func (h *TimedRotatingFileHandler) Close() error {
	return h.fd.Close()
}

func (h *TimedRotatingFileHandler) ListDir(dir, suffix string) (files []string, err error) {
	files = []string{}

	_dir, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range _dir {
		if file.IsDir() {
			// 不需要递归读取子目录下的文件
			continue
		}
		names := strings.Split(file.Name(), "_")
		if len(names) != 2 {
			// 文件名格式不对
			continue
		}
		_, err := time.Parse(h.suffix, names[0])
		if err != nil {
			// 文件名格式不对
			continue
		}

		if h.name != names[1] {
			// 文件名格式不对
			continue
		}
		files = append(files, filepath.Join(dir, file.Name()))
	}

	return files, nil
}
