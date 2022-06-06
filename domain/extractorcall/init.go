package extractorcall

import (
	"gorm.io/gorm"
)

type Config struct {
	RabbitMQConfig      MQConnectionConfig
	GetMetadataDatabase func() *gorm.DB
}

var globalMQManager *rabbitMQManager

const (
	QueueExtractorInput  = "extractor_input"
	QueueExtractorOutput = "extractor_output"
)

func Init(config *Config) {
	var err error
	globalMQManager, err = newRabbitMQManager(config.RabbitMQConfig.ToURL(), []string{
		QueueExtractorInput,
		QueueExtractorOutput,
	})
	if err != nil {
		panic(err)
	}

	err = globalMQManager.ListenOn(QueueExtractorOutput, buildReceive(config.GetMetadataDatabase))
	if err != nil {
		panic(err)
	}
}

func Close() {
	if globalMQManager != nil {
		err := globalMQManager.Close()
		if err != nil {
			globalMQManager.logger.WithError(err).Errorf("globalMQManager close fail with err:\n%v", err)
		}
	}
}
