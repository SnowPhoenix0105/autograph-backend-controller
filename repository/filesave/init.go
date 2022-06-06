package filesave

import (
	"fmt"
	"time"
)

type Config struct {
	Host    string
	Port    string
	TimeOut time.Duration
}

func (c *Config) FullHost() string {
	if len(c.Port) == 0 {
		return c.Host
	}
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func GenerateTestConfig() *Config {
	return &Config{
		Host:    "localhost",
		Port:    "8001",
		TimeOut: 10 * time.Second,
	}
}

var globalConfig Config

func Init(config *Config) {
	globalConfig = *config
}

func GetConfig() *Config {
	cpy := globalConfig
	return &cpy
}
