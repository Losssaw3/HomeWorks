package notifyconfig

import (
	"flag"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type NotificationConfig struct {
	Env            string `yaml:"env" env:"APP_ENV" env-default:"local"`
	StoragePath    string `yaml:"storage_path" env:"STORAGE_PATH" env-required:"true"`
	MigrationsPath string
	Kafka          KafkaConfig `yaml:"kafka"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers" env-default:"localhost:9092"`
	Topic   string   `yaml:"Topic" env-default:"default-topic"`
	GroupID string   `yaml:"group_id" env-default:"default-group"`
}

func MustLoad() *NotificationConfig {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg NotificationConfig

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("config path is empty: " + err.Error())
	}

	return &cfg

}

func fetchConfigPath() string {
	var res string
	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
