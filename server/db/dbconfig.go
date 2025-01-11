package db

import (
	"dt/utils"
	"io/fs"
	"strconv"
)

type DatabaseConfig struct {
	Host             string
	Port             int
	Username         string
	Password         string
	DBName           string
	Migrations       fs.FS
	MigrationsFolder string
}

type Option func(*DatabaseConfig)

func WithHost(host string) Option {
	return func(c *DatabaseConfig) {
		c.Host = host
	}
}

func WithPort(port int) Option {
	return func(c *DatabaseConfig) {
		c.Port = port
	}
}

func WithUsername(username string) Option {
	return func(c *DatabaseConfig) {
		c.Username = username
	}
}

func WithPassword(password string) Option {
	return func(c *DatabaseConfig) {
		c.Password = password
	}
}

func WithDBName(dbname string) Option {
	return func(c *DatabaseConfig) {
		c.DBName = dbname
	}
}

func WithMigrations(migrations fs.FS) Option {
	return func(c *DatabaseConfig) {
		c.Migrations = migrations
	}
}

func WithMigrationsFolder(folder string) Option {
	return func(c *DatabaseConfig) {
		c.MigrationsFolder = folder
	}
}

func LoadConfigFromEnv(migrations fs.FS, migrationsFolder string) *DatabaseConfig {
	port, _ := strconv.Atoi(utils.GetEnvOrDefault("DB_PORT", "5432"))

	return &DatabaseConfig{
		Host:             utils.GetEnvOrDefault("DB_HOST", "localhost"),
		Port:             port,
		Username:         utils.GetEnvOrDefault("DB_USER", "postgres"),
		Password:         utils.GetEnvOrDefault("DB_PASSWORD", "postgres"),
		DBName:           utils.GetEnvOrDefault("DB_NAME", "db-name"),
		Migrations:       migrations,
		MigrationsFolder: migrationsFolder,
	}
}

func NewDatabaseConfig(opts ...Option) *DatabaseConfig {
	dbConfig := &DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Username: "postgres",
		Password: "postgres",
		DBName:   "db-name",
	}
	for _, opt := range opts {
		opt(dbConfig)
	}
	return dbConfig
}
