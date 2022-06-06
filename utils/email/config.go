package email

type SMTPConfig struct {
	Identity string
	Host     string
	Port     int
	UserName string
	Password string
}

type Config struct {
	SMTP SMTPConfig
}

var globalConfig = Config{}

func Init(config *Config) {
	globalConfig = *config
}

func GenerateTestConfig() *Config {
	return &Config{SMTP: SMTPConfig{
		Identity: "autograph_sender@163.com",
		Host:     "smtp.163.com",
		Port:     25,
		UserName: "autograph_sender@163.com",
		Password: "FBIYWSQDFLKXRTKU",
	}}
}
