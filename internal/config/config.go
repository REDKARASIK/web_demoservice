package config

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
)

func NewConfig(path string) (*Config, error) {
	_, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("config file %s does not exist; err: %w", path, err)
	}

	var cfg Config
	if _, err = toml.DecodeFile(path, &cfg); err != nil {
		return nil, fmt.Errorf("error parsing config file %s; err: %w", path, err)
	}

	return &cfg, nil
}

type Config struct {
	HTTP  HTTPConfig     `toml:"http"`
	DB    PostgresConfig `toml:"db"`
	Kafka KafkaConfig    `toml:"kafka"`
}

type HTTPConfig struct {
	Host     string        `toml:"host"`
	Port     int           `toml:"port"`
	CacheTTL time.Duration `toml:"cache_ttl"`
}

type PostgresConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	SSLMode  string `toml:"sslmode"`

	MaxConns          int32         `toml:"max_conns"`
	MinConns          int32         `toml:"min_conns"`
	MaxConnLifetime   time.Duration `toml:"max_conn_lifetime"`
	HealthCheckPeriod time.Duration `toml:"health_check_period"`
}

func (p *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode,
	)
}

type KafkaConfig struct {
	Brokers  []string `toml:"brokers"`
	Topic    string   `toml:"topic"`
	GroupID  string   `toml:"group_id"`
	DLQTopic string   `toml:"dlq_topic"`
}
