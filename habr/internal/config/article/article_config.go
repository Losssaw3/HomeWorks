package articleconfig

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type ArticleConfig struct {
	Env            string        `yaml:"env" env:"APP_ENV" env-default:"local"`
	StoragePath    string        `yaml:"storage_path" env:"STORAGE_PATH" env-required:"true"`
	MigrationsPath string        `yaml:"migrations_path" env:"MIGRATIONS_PATH" env-required:"true"`
	HttpAddr       string        `yaml:"http_addr" env:"HTTP_ADDR" env-default:":8080"`
	Name           string        `yaml:"name" env:"APP_NAME" env-required:"true"`
	Secret         string        `yaml:"secret" env:"APP_SECRET" env-required:"true"`
	GRPCAddr       string        `yaml:"grpc_addr" env:"GRPC_ADDR" env-required:"true"`
	RedisAddr      string        `yaml:"redis_addr" env:"REDIS_ADDR" env-required:"true"`
	RedisPassword  string        `yaml:"redis_password" env:"REDIS_PASSWORD" env-default:""`
	RedisDBNum     int           `yaml:"redis_db_num" env:"REDIS_DB_NUM" env-default:"0"`
	RedisTTL       time.Duration `yaml:"redis_ttl" env:"REDIS_TTL" env-default:"5m"`
}

func MustLoad() *ArticleConfig {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg ArticleConfig

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
