package logging

import (
	"github.com/sirupsen/logrus"
	"os"
)

type consoleHook struct {
	level     logrus.Level
	formatter logrus.Formatter
}

func (h *consoleHook) Levels() []logrus.Level {
	return levels(h.level)
}

func (h *consoleHook) Fire(entry *logrus.Entry) error {
	b, _ := h.formatter.Format(entry)
	_, err := os.Stdout.Write(b)
	return err
}

func newConsoleHook(level logrus.Level, formatter logrus.Formatter) logrus.Hook {
	return &consoleHook{
		level:     level,
		formatter: formatter,
	}
}
