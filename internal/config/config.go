package config

import (
	"fmt"
	"time"
)

type Config struct {
	HTTP  HTTPConfig     `toml:"http"`
	DB    PostgresConfig `toml:"db"`
	Kafka KafkaConfig    `toml:"kafka"`
}

type HTTPConfig struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type PostgresConfig struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	SSLMode  string `toml:"sslmode"`

	MaxConns        int32         `toml:"max_conns"`
	MinConns        int32         `toml:"min_conns"`
	MaxConnLifetime time.Duration `toml:"max_conn_lifetime"`
}

func (p *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		p.User, p.Password, p.Host, p.Port, p.Database, p.SSLMode,
	)
}

type KafkaConfig struct {
	Brokers []string `toml:"brokers"`
	Topic   string   `toml:"topic"`
	GroupID string   `toml:"group_id"`
}
