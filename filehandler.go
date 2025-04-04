package simplelog

import (
	"fmt"
	"io"
	"os"
	"path"
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
	maxBytes    int
	curBytes    int
	backupCount int
}

func (h *RotatingFileHandler) InitRotating(name string, maxBytes int, backupCount int) (*RotatingFileHandler, error) {
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
		if f.Size() < int64(h.maxBytes) {
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
	h.curBytes = int(f.Size())

	return h, nil
}

func (h *RotatingFileHandler) Write(p []byte) (n int, err error) {
	h.doRollover()
	n, err = h.fd.Write(p)
	h.curBytes += n
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
		h.curBytes = int(f.Size())
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

		h.fd, _ = os.OpenFile(h.fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		h.curBytes = 0
		f, err := h.fd.Stat()
		if err != nil {
			return
		}
		h.curBytes = int(f.Size())
	}
}

// TimedRotatingFileHandler writes log to a file,
// it will backup current and open a new one, with a period time you sepecified.
//
// refer: http://docs.python.org/2/library/logging.handlers.html.
// same like python TimedRotatingFileHandler.
type TimedRotatingFileHandler struct {
	fd *os.File

	baseName   string
	interval   int64
	suffix     string
	rolloverAt int64
}

const (
	WhenSecond = iota
	WhenMinute
	WhenHour
	WhenDay
)

func (h *TimedRotatingFileHandler) InitTimedRotating(name string, when int8, interval int) (*TimedRotatingFileHandler, error) {
	dir := path.Dir(name)
	os.Mkdir(dir, 0777)

	h.baseName = name

	switch when {
	case WhenSecond:
		h.interval = 1
		h.suffix = "2006-01-02_15-04-05"
	case WhenMinute:
		h.interval = 60
		h.suffix = "2006-01-02_15-04"
	case WhenHour:
		h.interval = 3600
		h.suffix = "2006-01-02_15"
	case WhenDay:
		h.interval = 3600 * 24
		h.suffix = "2006-01-02"
	default:
		return nil, fmt.Errorf("invalid when_rotate: %d", when)
	}

	h.interval = h.interval * int64(interval)

	var err error
	h.fd, err = os.OpenFile(h.baseName, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	fInfo, _ := h.fd.Stat()
	h.rolloverAt = fInfo.ModTime().Unix() + h.interval

	return h, nil
}

func (h *TimedRotatingFileHandler) doRollover() {
	//refer http://hg.python.org/cpython/file/2.7/Lib/logging/handlers.py
	now := time.Now()

	if h.rolloverAt <= now.Unix() {
		fName := h.baseName + now.Format(h.suffix)
		h.fd.Close()
		e := os.Rename(h.baseName, fName)
		if e != nil {
			panic(e)
		}

		h.fd, _ = os.OpenFile(h.baseName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)

		h.rolloverAt = time.Now().Unix() + h.interval
	}
}

func (h *TimedRotatingFileHandler) Write(b []byte) (n int, err error) {
	h.doRollover()
	return h.fd.Write(b)
}

func (h *TimedRotatingFileHandler) Close() error {
	return h.fd.Close()
}
