package neograph

import (
	"autograph-backend-controller/utils"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Neo4jConfig struct {
	Host string
	Port int
	User string
	Pwd  string
}

type Config struct {
	Neo4j Neo4jConfig
}

func GenerateTestConfig() *Config {
	return &Config{Neo4j: Neo4jConfig{
		Host: "localhost",
		Port: 7687,
		User: "neo4j",
		Pwd:  "autograph",
	}}
}

var globalDriver neo4j.Driver
var globalConfig Config

func initDriver(neoConfig *Neo4jConfig) (neo4j.Driver, error) {
	url := fmt.Sprintf("%s://%s:%d", "neo4j", neoConfig.Host, neoConfig.Port)

	ret, err := neo4j.NewDriver(url, neo4j.BasicAuth(neoConfig.User, neoConfig.Pwd, ""))
	if err != nil {
		return nil, utils.WrapErrorf(err, "init neo4j.Driver with config [%#v] fail", neoConfig)
	}

	return ret, nil
}

func Init(config *Config) {
	var err error

	globalConfig = *config
	globalDriver, err = initDriver(&config.Neo4j)
	if err != nil {
		panic(err)
	}
}

func Close() {
	globalDriver.Close()
}
