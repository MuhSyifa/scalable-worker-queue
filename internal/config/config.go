package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Postgres PostgresConfig `mapstructure:"postgres"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Worker   WorkerConfig   `mapstructure:"worker"`
}

type AppConfig struct {
	Name        string `mapstructure:"name"`
	Environment string `mapstructure:"environment"`
	Port        string `mapstructure:"port"`
}

type PostgresConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type WorkerConfig struct {
	Concurrency int `mapstructure:"concurrency"`
	MaxRetries  int `mapstructure:"max_retries"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Default values
	viper.SetDefault("app.name", "golang-worker-queue")
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.port", "8080")

	viper.SetDefault("postgres.host", "localhost")
	viper.SetDefault("postgres.port", 5432)
	viper.SetDefault("postgres.user", "postgres")
	viper.SetDefault("postgres.password", "@kurakura06")
	viper.SetDefault("postgres.dbname", "worker_queue")
	viper.SetDefault("postgres.sslmode", "disable")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("worker.concurrency", 10)
	viper.SetDefault("worker.max_retries", 3)

	_ = viper.ReadInConfig() // Ignore error if .env file is not found, use env vars/defaults

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Printf("Unable to decode into struct, %v", err)
		return nil, err
	}

	return &cfg, nil
}
