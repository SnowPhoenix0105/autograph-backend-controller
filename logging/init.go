package logging

import (
	"autograph-backend-controller/config"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type Config struct {
	FileLevel      logrus.Level
	ConsoleLevel   logrus.Level
	FileDir        string
	DisableConsole bool
	Testing        *testing.T
}

var defaultConfig = Config{
	ConsoleLevel:   logrus.DebugLevel,
	DisableConsole: false,
}

func GenerateTestConfig(t *testing.T) *Config {
	fullConsoleLog := os.Getenv(config.EnvKeyTestFullConsoleLog)
	consoleLevel := logrus.InfoLevel
	if len(fullConsoleLog) != 0 {
		consoleLevel = logrus.DebugLevel
	}

	fileDir := os.Getenv(config.EnvKeyTestLogDir)
	if len(fileDir) == 0 {
		projectName := "autograph-backend-controller"
		pwd, _ := os.Getwd()
		abs, _ := filepath.Abs(pwd)
		base := abs[:strings.LastIndex(abs, projectName)+len(projectName)]
		fileDir = filepath.Join(base, "logs")
	}
	return &Config{
		FileLevel:      logrus.DebugLevel,
		ConsoleLevel:   consoleLevel,
		FileDir:        fileDir,
		DisableConsole: false,
		Testing:        t,
	}
}

var defaultLogger = NewLogger()

func SetDefaultConfig(cfg *Config) {
	defaultConfig = *cfg
	defaultLogger = NewLogger()
}

func Default() *logrus.Logger {
	return defaultLogger
}

func configOutput(logger *logrus.Logger, cfg *Config, formatter logrus.Formatter) {
	logger.Out = io.Discard
	if len(cfg.FileDir) != 0 {
		fileHook, err := newFileHook(cfg.FileDir, cfg.FileLevel, formatter)
		if err != nil {
			panic(err)
		}
		if cfg.DisableConsole {
			logger.Out = fileHook
			return
		}
		logger.AddHook(fileHook)
	}
	if !cfg.DisableConsole {
		if cfg.Testing != nil {
			logger.AddHook(newTestHook(cfg.Testing, cfg.ConsoleLevel, formatter))
			return
		}
		logger.AddHook(newConsoleHook(cfg.ConsoleLevel, formatter))
	}
}

func NewLoggerWithConfig(cfg *Config) *logrus.Logger {
	ret := logrus.New()

	// 配置日志格式
	ret.ReportCaller = true
	formatter := logFormatter{
		timeLayout: "2006-01-02 15:04:05.000",
	}
	ret.SetFormatter(&formatter)

	// 配置日志等级
	minLevel := cfg.FileLevel
	if len(cfg.FileDir) == 0 || cfg.ConsoleLevel > minLevel {
		// 等级越高 Level 的值越小
		minLevel = cfg.ConsoleLevel
	}
	ret.SetLevel(minLevel)

	// 配置日志输出
	configOutput(ret, cfg, &formatter)

	return ret
}

func NewLogger() *logrus.Logger {
	return NewLoggerWithConfig(&defaultConfig)
}
