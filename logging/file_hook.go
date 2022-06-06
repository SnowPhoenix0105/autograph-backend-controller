package logging

import (
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

type fileHook struct {
	level     logrus.Level
	fileTime  time.Time
	file      *os.File
	dir       string
	formatter logrus.Formatter
	mutex     sync.Mutex
}

func (h *fileHook) updateFile(now time.Time) error {
	if h.fileTime.Day() != now.Day() {
		h.mutex.Lock()
		if h.fileTime.Day() != now.Day() {
			f, err := openLogFile(h.dir, now)
			if err != nil {
				h.mutex.Unlock()
				return err
			}
			h.file = f
			h.fileTime = now
		}
		h.mutex.Unlock()
	}
	return nil
}

func (h *fileHook) Write(p []byte) (n int, err error) {
	_ = h.updateFile(time.Now())
	n, err = h.file.Write(p)
	return
}

func (h *fileHook) Levels() []logrus.Level {
	return levels(h.level)
}

func (h *fileHook) Fire(entry *logrus.Entry) error {
	_ = h.updateFile(entry.Time)
	b, _ := h.formatter.Format(entry)
	_, err := h.file.Write(b)
	return err
}

func newFileHook(dir string, level logrus.Level, formatter logrus.Formatter) (*fileHook, error) {
	now := time.Now()
	f, err := openLogFile(dir, now)
	if err != nil {
		return nil, err
	}
	return &fileHook{
		level:     level,
		fileTime:  now,
		file:      f,
		dir:       dir,
		formatter: formatter,
		mutex:     sync.Mutex{},
	}, nil
}
