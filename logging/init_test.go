package logging

import (
	"autograph-backend-controller/config"
	"autograph-backend-controller/utils"
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestDouble(t *testing.T) {
	logDir := os.Getenv(config.EnvKeyTestLogDir)
	t.Logf("%#v", logDir)

	logger := NewLoggerWithConfig(&Config{
		FileLevel:      logrus.DebugLevel,
		ConsoleLevel:   logrus.InfoLevel,
		FileDir:        logDir,
		DisableConsole: false,
	})

	logger.Tracef("trace")
	logger.Debugf("debug")
	logger.Infof("info")
	logger.Warnf("warn")
	logger.Errorf("error")
	func(t2 *testing.T) {
		defer func() {
			r := recover()
			assert.NotNil(t2, r)
		}()
		logger.Panic("panic")
	}(t)
}

func TestFileOnly(t *testing.T) {
	logger := NewLoggerWithConfig(&Config{
		FileLevel:      logrus.DebugLevel,
		ConsoleLevel:   logrus.InfoLevel,
		FileDir:        os.Getenv(config.EnvKeyTestLogDir),
		DisableConsole: true,
	})
	logger.Tracef("trace")
	logger.Debugf("debug")
	logger.Infof("info")
	logger.Warnf("warn")
	logger.Errorf("error")
}

func TestConsoleOnly(t *testing.T) {
	logger := NewLoggerWithConfig(&Config{
		FileLevel:      logrus.DebugLevel,
		ConsoleLevel:   logrus.InfoLevel,
		FileDir:        "",
		DisableConsole: false,
	})
	logger.Tracef("trace")
	logger.Debugf("debug")
	logger.Infof("info")
	logger.Warnf("warn")
	logger.Errorf("error")
}

func TestError(t *testing.T) {
	SetDefaultConfig(GenerateTestConfig(t))
	logger := NewLogger()
	err := errors.New("TestError")
	err = utils.WrapError(err, "wrapped1 err")
	err = utils.WrapError(err, "wrapped2 err")
	logger.WithError(err).Error("Test LoggerWithError")
	logger.WithError(err).Errorf("Test LoggerWithError:\n%s", err)
	logger.WithError(err).Errorf("Test LoggerWithError:\n%v", err)
	logger.WithError(err).Errorf("Test LoggerWithError:\n%+v", err)
	logger.WithError(err).Errorf("Test LoggerWithError:\n%#v", err)
}
