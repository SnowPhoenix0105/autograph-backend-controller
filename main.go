package main

import (
	"autograph-backend-controller/config"
	"autograph-backend-controller/domain/extractorcall"
	"autograph-backend-controller/domain/graph"
	"autograph-backend-controller/domain/tagger"
	"autograph-backend-controller/logging"
	"autograph-backend-controller/repository/filesave"
	"autograph-backend-controller/repository/metadata"
	"autograph-backend-controller/repository/neograph"
	"autograph-backend-controller/server"
	"autograph-backend-controller/utils"
	"autograph-backend-controller/utils/email"
	"github.com/sirupsen/logrus"
	"os"
)

const DEBUG = true

func loggingConf() *logging.Config {
	return &logging.Config{
		FileLevel:      logrus.DebugLevel,
		ConsoleLevel:   logrus.InfoLevel,
		FileDir:        "logs",
		DisableConsole: false,
	}
}

func emailConf() *email.Config {
	if DEBUG {
		return email.GenerateTestConfig()
	}

	return &email.Config{SMTP: email.SMTPConfig{
		Identity: os.Getenv(config.EnvKeyEmailSMTPIdentity),
		Host:     os.Getenv(config.EnvKeyEmailSMTPHost),
		Port:     utils.MustAtoi(os.Getenv(config.EnvKeyEmailSMTPPort)),
		UserName: os.Getenv(config.EnvKeyEmailSMTPUserName),
		Password: os.Getenv(config.EnvKeyEmailSMTPPassword),
	}}
}

func metadataConf() *metadata.Config {
	if DEBUG {
		return metadata.GenerateTestConfig()
	}

	// TODO
	return &metadata.Config{
		MySQL:          metadata.MySQLConfig{},
		CheckMigration: false,
	}
}

func taggerConf() *tagger.TagSetting {
	return &tagger.TagSetting{
		Logger:              logging.NewLogger(),
		GetMetadataDatabase: metadata.DatabaseRaw,
	}
}

func filesaveConf() *filesave.Config {
	if DEBUG {
		return filesave.GenerateTestConfig()
	}

	// TODO
	return &filesave.Config{
		Host:    "",
		Port:    "",
		TimeOut: 0,
	}
}

func extractorcallConf() *extractorcall.Config {
	if DEBUG {
		return &extractorcall.Config{
			RabbitMQConfig:      extractorcall.GenerateTestMQConnectionConfig(),
			GetMetadataDatabase: metadata.DatabaseRaw,
		}
	}

	// TODO
	return &extractorcall.Config{
		RabbitMQConfig:      extractorcall.MQConnectionConfig{},
		GetMetadataDatabase: metadata.DatabaseRaw,
	}
}

func neographConf() *neograph.Config {
	if DEBUG {
		return neograph.GenerateTestConfig()
	}

	// TODO
	return &neograph.Config{Neo4j: neograph.Neo4jConfig{
		Host: "",
		Port: 0,
		User: "",
		Pwd:  "",
	}}
}

func graphConf() *graph.KGSetting {
	return &graph.KGSetting{
		GetMetadataDatabase: metadata.DatabaseRaw,
		Logger:              logging.NewLogger(),
	}
}

func main() {
	logging.SetDefaultConfig(loggingConf())
	logger := logging.NewLogger()

	email.Init(emailConf())

	metadata.Init(metadataConf())

	tagger.Init(taggerConf())

	filesave.Init(filesaveConf())

	extractorcall.Init(extractorcallConf())
	defer extractorcall.Close()

	neograph.Init(neographConf())
	defer neograph.Close()

	graph.Init(graphConf())

	s := server.New(&server.Config{
		Host:      "",
		Port:      8003,
		DebugMode: DEBUG,
	})
	err := s.RunServer()
	if err != nil {
		logger.WithError(err).Errorf("run server error=\n%v", err)
	}
}
