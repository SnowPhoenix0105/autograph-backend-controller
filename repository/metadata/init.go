package metadata

import (
	"autograph-backend-controller/logging"
	"autograph-backend-controller/utils"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type MySQLConfig struct {
	User     string
	Password string
	Host     string
	Database string
}

func (c *MySQLConfig) dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Database)
}

type Config struct {
	MySQL          MySQLConfig
	CheckMigration bool
}

func GenerateTestConfig() *Config {
	return &Config{
		MySQL: MySQLConfig{
			User:     "metadata_test",
			Password: "metadata_test",
			Host:     "localhost",
			Database: "metadata_test",
		},
		CheckMigration: true,
	}
}

var db *gorm.DB

func CreateDatabase(config *Config) (*gorm.DB, error) {
	database, err := gorm.Open(mysql.Open(config.MySQL.dsn()), &gorm.Config{
		Logger: logger.New(&sqlLogger{logger: logging.NewLogger()}, logger.Config{LogLevel: logger.Info}),
	})
	if err != nil {
		return nil, utils.WrapError(err, "db connection fail")
	}

	if config.CheckMigration {
		err = migration(database)
		if err != nil {
			return nil, utils.WrapError(err, "migration fail")
		}

		err = ensureDemo(database)
		if err != nil {
			return nil, utils.WrapError(err, "ensureDemo fail")
		}
	}

	return database, nil
}

func ensureDemo(db *gorm.DB) error {

	demoExtractor := Extractor{
		Name: "demoextractor",
		Type: ExtractorTypeModel,
		Desc: "demo",
	}
	err := db.FirstOrCreate(&demoExtractor).Error

	if err != nil {
		return utils.WrapError(err, "first or create demo fail")
	}

	if demoExtractor.ID != 1 {
		return fmt.Errorf("demo_extractor as id=[%d]", demoExtractor.ID)
	}

	return nil
}

func migration(db *gorm.DB) error {
	tables := []interface{}{
		&File{}, &Text{}, &Extractor{},
		&ExtractTask{}, &ExtractTaskItem{},
		&Relation{}, &Entity{},
		&Build{}, &BuildExtractor{}, &Node{},
	}
	err := db.
		Set("gorm:table_options", "ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci").
		AutoMigrate(tables...)
	if err != nil {
		return utils.WrapError(err, "AutoMigrate fail")
	}

	return nil
}

func Init(config *Config) {
	database, err := CreateDatabase(config)
	if err != nil {
		panic(err)
	}

	db = database
}

func DatabaseRaw() *gorm.DB {
	return db
}
