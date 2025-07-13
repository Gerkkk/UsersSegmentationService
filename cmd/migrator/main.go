package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"os"
	"strings"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// ShardConfig - структура для парсинга конфига одного шарда.
type ShardConfig struct {
	Name   string `mapstructure:"name"`
	DSNEnv string `mapstructure:"dsn_env"`
}

// Config - структура для парсинга конфига мигратора
type Config struct {
	DB struct {
		NumShards int           `mapstructure:"num_shards"`
		Shards    []ShardConfig `mapstructure:"shards"`
	} `mapstructure:"db"`
	MigrationsPath  string `mapstructure:"migrations_path"`
	MigrationsTable string `mapstructure:"migrations_table"`
}

// loadConfig - функция загрузки конфига мигратора
func loadConfig() (*Config, error) {
	var cfgPath string

	flag.StringVar(&cfgPath, "config", "", "path to config file")
	flag.Parse()

	if cfgPath == "" {
		cfgPath = os.Getenv("CONFIG_PATH")
	}

	viper.SetConfigFile(cfgPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}

func main() {
	if err := godotenv.Load(); err != nil {
		panic("Error loading .env file: " + err.Error())
	}

	cfg, err := loadConfig()
	if err != nil {
		panic("Error loading config file: " + err.Error())
	}

	if cfg.MigrationsPath == "" {
		panic("migrations_path must be set in config")
	}
	if cfg.MigrationsTable == "" {
		cfg.MigrationsTable = "schema_migrations"
	}

	for _, shard := range cfg.DB.Shards {
		dsn := os.Getenv(shard.DSNEnv)
		if dsn == "" {
			panic("Missing DSN for shard " + shard.Name + " (env: " + shard.DSNEnv + ")")
		}

		fmt.Printf("Migrating shard: %s (%s)\n", shard.Name, shard.DSNEnv)

		separator := "?"
		if strings.Contains(dsn, "?") {
			separator = "&"
		}
		dbURL := fmt.Sprintf("%s%sx-migrations-table=%s", dsn, separator, cfg.MigrationsTable)

		m, err := migrate.New(
			"file://"+cfg.MigrationsPath,
			dbURL,
		)
		if err != nil {
			panic("failed to create migration for " + shard.Name + ": " + err.Error())
		}

		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Printf("No migrations to apply for %s\n", shard.Name)
				continue
			}
			panic("failed to apply migration for " + shard.Name + ": " + err.Error())
		}

		fmt.Printf("Migrations applied for shard: %s\n", shard.Name)
	}
}
