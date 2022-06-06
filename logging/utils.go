package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"time"
)

func openLogFile(dir string, now time.Time) (*os.File, error) {
	p := fmt.Sprintf("%d-%d-%d.log", now.Year(), now.Month(), now.Day())
	p = path.Join(dir, p)
	return os.OpenFile(p, os.O_CREATE|os.O_APPEND|os.O_WRONLY, os.ModeAppend|os.ModePerm)
}

func levels(top logrus.Level) []logrus.Level {
	var ret []logrus.Level
	for l := logrus.PanicLevel; l <= top; l++ {
		ret = append(ret, l)
	}
	return ret
}
