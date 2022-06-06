package metadata

import (
	"github.com/sirupsen/logrus"
)

type sqlLogger struct {
	logger *logrus.Logger
}

func (l *sqlLogger) Printf(fmt string, args ...interface{}) {
	l.logger.Debugf(fmt, args...)
}
