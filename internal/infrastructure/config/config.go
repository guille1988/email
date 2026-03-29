package config

import (
	"email/internal/infrastructure/env"

	"github.com/joho/godotenv"
)

// Config represents the application configuration.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Log      LogConfig
	RabbitMQ RabbitMQConfig
	Mail     MailConfig
}

type MailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

// AppConfig represents the application configuration.
type AppConfig struct {
	Name string
	Env  Env
}

type ConnectionName string

const (
	Default ConnectionName = "default"
)

// DatabaseConfig represents the database configuration.
type DatabaseConfig struct {
	Connections map[ConnectionName]DatabaseConnection
}

// DatabaseConnection represents the database connection.
type DatabaseConnection struct {
	Driver             Driver
	Host               string
	Port               string
	Database           string
	Username           string
	Password           string
	MaxIdleConnections int
	MaxOpenConnections int
}

type LogConfig struct {
	Driver LogDriver
	Path   string
	Level  LogLevel
}

type Driver string

const (
	MySQLDriver    Driver = "mysql"
	PostgresDriver Driver = "postgres"
	Sqlite         Driver = "sqlite"
)

type Env string

const (
	LocalEnv      Env = "local"
	TestingEnv    Env = "testing"
	StagingEnv    Env = "staging"
	ProductionEnv Env = "production"
)

type LogLevel string

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
)

type LogDriver string

const (
	StdoutFormat LogDriver = "stdout"
	File         LogDriver = "file"
)

// New creates a new configuration instance.
func New() (*Config, error) {
	_ = godotenv.Load()

	config := Config{
		App: AppConfig{
			Name: env.GetEnvAsString("APP_NAME", "email"),
			Env:  Env(env.GetEnvAsString("APP_ENV", string(LocalEnv))),
		},
		Database: DatabaseConfig{
			Connections: map[ConnectionName]DatabaseConnection{
				Default: {
					Driver:             Driver(env.GetEnvAsString("DB_DRIVER", string(MySQLDriver))),
					Host:               env.GetEnvAsString("DB_HOST", "mysql_email"),
					Port:               env.GetEnvAsString("DB_PORT", "3307"),
					Database:           env.GetEnvAsString("DB_DATABASE", "email"),
					Username:           env.GetEnvAsString("DB_USERNAME", "email"),
					Password:           env.GetEnvAsString("DB_PASSWORD", "email"),
					MaxIdleConnections: env.GetEnvAsInt("DB_MAX_IDLE_CONNECTIONS", 10),
					MaxOpenConnections: env.GetEnvAsInt("DB_MAX_OPEN_CONNECTIONS", 10),
				},
			},
		},
		Log: LogConfig{
			Driver: LogDriver(env.GetEnvAsString("LOG_DRIVER", string(StdoutFormat))),
			Path:   env.GetEnvAsString("LOG_PATH", "logs/email.log"),
			Level:  LogLevel(env.GetEnvAsString("LOG_LEVEL", string(InfoLevel))),
		},
		RabbitMQ: RabbitMQConfig{
			Host:     env.GetEnvAsString("RABBITMQ_HOST", "rabbitmq"),
			Port:     env.GetEnvAsString("RABBITMQ_PORT", "5672"),
			User:     env.GetEnvAsString("RABBITMQ_USER", "guest"),
			Password: env.GetEnvAsString("RABBITMQ_PASSWORD", "guest"),
		},
		Mail: MailConfig{
			Host:     env.GetEnvAsString("MAIL_HOST", "mailpit"),
			Port:     env.GetEnvAsInt("MAIL_PORT", 1025),
			User:     env.GetEnvAsString("MAIL_USER", ""),
			Password: env.GetEnvAsString("MAIL_PASSWORD", ""),
		},
	}

	return &config, nil
}
