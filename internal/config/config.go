package config

import (
	"flag"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env   string      `yaml:"env" env-default:"local"`
	Grpc  GrpcConfig  `yaml:"grpc"`
	Db    DbConfig    `yaml:"db"`
	Cache CacheConfig `yaml:"cache"`
	Queue QueueConfig `yaml:"queue"`
}

type QueueConfig struct {
	Brokers []string `yaml:"brokers"`
	Topics  []string `yaml:"topics"`
	Group   string   `yaml:"group"`
}

type GrpcConfig struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DbConfig struct {
	NumShards int           `yaml:"num_shards"`
	Shards    []ShardConfig `yaml:"shards"`
}

type CacheConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	Db              string `yaml:"db"`
	MaxMemory       string `yaml:"max_memory"`
	MaxMemoryPolicy string `yaml:"max_memory_policy"`
	UserSegmentsTtl string `yaml:"user_segments_ttl"`
	PasswordEnv     string `yaml:"password_env"`
	Password        string
}

type ShardConfig struct {
	Name   string `yaml:"name"`
	DSNEnv string `yaml:"dsn_env"`
	DSN    string
}

func MustLoadConfig() *Config {
	path := fetchConfigPath()

	if path == "" {
		panic("config path is empty")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config path does not exist: " + path)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic("Error reading configs: " + err.Error())
	}

	for i := range cfg.Db.Shards {
		DSN := os.Getenv(cfg.Db.Shards[i].DSNEnv)

		if DSN == "" {
			panic("env variable " + cfg.Db.Shards[i].DSNEnv + " is not set")
		}

		cfg.Db.Shards[i].DSN = DSN
	}

	cachePwd := os.Getenv(cfg.Cache.PasswordEnv)

	if cachePwd == "" {
		panic("env variable cache password is not set")
	}

	cfg.Cache.Password = cachePwd

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
