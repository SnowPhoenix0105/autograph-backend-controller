package logging

import (
	"github.com/sirupsen/logrus"
	"testing"
)

type testHook struct {
	Testing   *testing.T
	level     logrus.Level
	formatter logrus.Formatter
}

func (t *testHook) Write(p []byte) (n int, err error) {
	t.Testing.Log("\n" + string(p))
	return len(p), nil
}

func (t *testHook) Levels() []logrus.Level {
	return levels(t.level)
}

func (t *testHook) Fire(entry *logrus.Entry) error {
	b, _ := t.formatter.Format(entry)
	_, err := t.Write(b)
	return err
}

func newTestHook(t *testing.T, level logrus.Level, formatter logrus.Formatter) logrus.Hook {
	return &testHook{
		Testing:   t,
		level:     level,
		formatter: formatter,
	}
}
