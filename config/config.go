package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ilhamosaurus/sns-platform/pkg/db"
	"gopkg.in/yaml.v3"
)

// AppConfig represents the entire application configuration
type AppConfig struct {
	Database   DatabaseConfig  `yaml:"database"`
	Postgres   PostgresConfig  `yaml:"postgres"`
	MySQL      MySQLConfig     `yaml:"mysql"`
	SQLite     SQLiteConfig    `yaml:"sqlite"`
	Redis      RedisConfig     `yaml:"redis"`
	App        ApplicationInfo `yaml:"app"`
	Migrations MigrationConfig `yaml:"migrations"`

	// Environment-specific configs
	Development *EnvironmentConfig `yaml:"development,omitempty"`
	Testing     *EnvironmentConfig `yaml:"testing,omitempty"`
	Staging     *EnvironmentConfig `yaml:"staging,omitempty"`
	Production  *EnvironmentConfig `yaml:"production,omitempty"`
}

// DatabaseConfig holds common database settings
type DatabaseConfig struct {
	Type            string        `yaml:"type"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	LogLevel        string        `yaml:"log_level"`
	PrepareStmt     bool          `yaml:"prepare_stmt"`
	SkipDefaultTxn  bool          `yaml:"skip_default_txn"`
}

// PostgresConfig holds PostgreSQL-specific settings
type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

// MySQLConfig holds MySQL-specific settings
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	Charset  string `yaml:"charset"`
}

// SQLiteConfig holds SQLite-specific settings
type SQLiteConfig struct {
	FilePath string `yaml:"filepath"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Enable       bool   `yaml:"enable"`
	Host         string `yaml:"host"`
	Port         string `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

// ApplicationInfo holds application metadata
type ApplicationInfo struct {
	Name        string          `yaml:"name"`
	Version     string          `yaml:"version"`
	Environment string          `yaml:"environment"`
	Port        int             `yaml:"port"`
	Features    map[string]bool `yaml:"features"`
}

// MigrationConfig holds migration settings
type MigrationConfig struct {
	AutoMigrate   bool `yaml:"auto_migrate"`
	SeedData      bool `yaml:"seed_data"`
	CreateIndexes bool `yaml:"create_indexes"`
}

// EnvironmentConfig holds environment-specific overrides
type EnvironmentConfig struct {
	Database DatabaseConfig `yaml:"database"`
	Postgres PostgresConfig `yaml:"postgres"`
	MySQL    MySQLConfig    `yaml:"mysql"`
	SQLite   SQLiteConfig   `yaml:"sqlite"`
	LogLevel string         `yaml:"log_level"`
}

var Config *AppConfig

// Load loads configuration from YAML file and environment variables
func Load(configPath string) (*AppConfig, error) {
	// Read YAML file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config AppConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Get environment (from env var or config)
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = config.App.Environment
	}
	if env == "" {
		env = "development"
	}

	// Apply environment-specific overrides
	if err := applyEnvironmentOverrides(&config, env); err != nil {
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	// Override with environment variables
	if err := overrideWithEnvVars(&config); err != nil {
		return nil, fmt.Errorf("failed to override with environment variables: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	Config = &config
	return &config, nil
}

// applyEnvironmentOverrides applies environment-specific settings
func applyEnvironmentOverrides(config *AppConfig, env string) error {
	var envConfig *EnvironmentConfig

	switch strings.ToLower(env) {
	case "development":
		envConfig = config.Development
	case "testing":
		envConfig = config.Testing
	case "staging":
		envConfig = config.Staging
	case "production":
		envConfig = config.Production
	default:
		return nil // No override
	}

	if envConfig == nil {
		return nil
	}

	// Override database settings
	if envConfig.Database.Type != "" {
		config.Database.Type = envConfig.Database.Type
	}
	if envConfig.Database.MaxIdleConns > 0 {
		config.Database.MaxIdleConns = envConfig.Database.MaxIdleConns
	}
	if envConfig.Database.MaxOpenConns > 0 {
		config.Database.MaxOpenConns = envConfig.Database.MaxOpenConns
	}
	if envConfig.LogLevel != "" {
		config.Database.LogLevel = envConfig.LogLevel
	}

	// Override database-specific settings
	if envConfig.Postgres.Host != "" {
		config.Postgres = envConfig.Postgres
	}
	if envConfig.MySQL.Host != "" {
		config.MySQL = envConfig.MySQL
	}
	if envConfig.SQLite.FilePath != "" {
		config.SQLite = envConfig.SQLite
	}

	return nil
}

// overrideWithEnvVars overrides config with environment variables
func overrideWithEnvVars(config *AppConfig) error {
	// Database type
	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		config.Database.Type = dbType
	}

	// PostgreSQL
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Postgres.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		config.Postgres.Port = port
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.Postgres.User = user
		config.MySQL.User = user // Apply to MySQL too
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Postgres.Password = password
		config.MySQL.Password = password
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Postgres.DBName = dbName
		config.MySQL.DBName = dbName
	}
	if sslMode := os.Getenv("DB_SSLMODE"); sslMode != "" {
		config.Postgres.SSLMode = sslMode
	}

	// MySQL specific
	if charset := os.Getenv("DB_CHARSET"); charset != "" {
		config.MySQL.Charset = charset
	}

	// SQLite
	if filepath := os.Getenv("DB_FILEPATH"); filepath != "" {
		config.SQLite.FilePath = filepath
	}

	// Redis
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		config.Redis.Host = redisHost
	}
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		config.Redis.Port = redisPort
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		config.Redis.Password = redisPassword
	}

	// Application
	if appPort := os.Getenv("APP_PORT"); appPort != "" {
		fmt.Sscanf(appPort, "%d", &config.App.Port)
	}

	return nil
}

// validateConfig validates the configuration
func validateConfig(config *AppConfig) error {
	// Validate database type
	dbType := db.DatabaseType(config.Database.Type)
	switch dbType {
	case db.PostgreSQL:
		if config.Postgres.Host == "" || config.Postgres.Port == "" {
			return fmt.Errorf("PostgreSQL configuration is incomplete")
		}
	case db.MySQL:
		if config.MySQL.Host == "" || config.MySQL.Port == "" {
			return fmt.Errorf("MySQL configuration is incomplete")
		}
	case db.SQLite:
		if config.SQLite.FilePath == "" {
			return fmt.Errorf("SQLite file path is required")
		}
	default:
		return fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}

	return nil
}

// GetDatabaseConfig converts AppConfig to database.Config
func (c *AppConfig) GetDatabaseConfig() db.Config {
	dbConfig := db.Config{
		Type:            db.DatabaseType(c.Database.Type),
		MaxIdleConns:    c.Database.MaxIdleConns,
		MaxOpenConns:    c.Database.MaxOpenConns,
		ConnMaxLifetime: c.Database.ConnMaxLifetime,
		ConnMaxIdleTime: c.Database.ConnMaxIdleTime,
		LogLevel:        c.Database.LogLevel,
		PrepareStmt:     c.Database.PrepareStmt,
		SkipDefaultTxn:  c.Database.SkipDefaultTxn,
	}

	// Set database-specific configs
	switch dbConfig.Type {
	case db.PostgreSQL:
		dbConfig.Host = c.Postgres.Host
		dbConfig.Port = c.Postgres.Port
		dbConfig.User = c.Postgres.User
		dbConfig.Password = c.Postgres.Password
		dbConfig.DBName = c.Postgres.DBName
		dbConfig.SSLMode = c.Postgres.SSLMode
	case db.MySQL:
		dbConfig.Host = c.MySQL.Host
		dbConfig.Port = c.MySQL.Port
		dbConfig.User = c.MySQL.User
		dbConfig.Password = c.MySQL.Password
		dbConfig.DBName = c.MySQL.DBName
		dbConfig.Charset = c.MySQL.Charset
	case db.SQLite:
		dbConfig.FilePath = c.SQLite.FilePath
	}

	return dbConfig
}

// PrintConfig prints the current configuration (safe for logging)
func (c *AppConfig) PrintConfig() {
	fmt.Println("=== Application Configuration ===")
	fmt.Printf("App Name: %s\n", c.App.Name)
	fmt.Printf("Version: %s\n", c.App.Version)
	fmt.Printf("Environment: %s\n", c.App.Environment)
	fmt.Printf("Port: %d\n", c.App.Port)
	fmt.Println()

	fmt.Println("=== Database Configuration ===")
	fmt.Printf("Type: %s\n", c.Database.Type)
	fmt.Printf("Max Idle Conns: %d\n", c.Database.MaxIdleConns)
	fmt.Printf("Max Open Conns: %d\n", c.Database.MaxOpenConns)
	fmt.Printf("Log Level: %s\n", c.Database.LogLevel)

	switch db.DatabaseType(c.Database.Type) {
	case db.PostgreSQL:
		fmt.Printf("Host: %s:%s\n", c.Postgres.Host, c.Postgres.Port)
		fmt.Printf("Database: %s\n", c.Postgres.DBName)
		fmt.Printf("User: %s\n", c.Postgres.User)
		fmt.Printf("SSL Mode: %s\n", c.Postgres.SSLMode)
	case db.MySQL:
		fmt.Printf("Host: %s:%s\n", c.MySQL.Host, c.MySQL.Port)
		fmt.Printf("Database: %s\n", c.MySQL.DBName)
		fmt.Printf("User: %s\n", c.MySQL.User)
		fmt.Printf("Charset: %s\n", c.MySQL.Charset)
	case db.SQLite:
		fmt.Printf("File: %s\n", c.SQLite.FilePath)
	}

	fmt.Println()
	fmt.Println("=== Redis Configuration ===")
	fmt.Printf("Host: %s:%s\n", c.Redis.Host, c.Redis.Port)
	fmt.Printf("Database: %d\n", c.Redis.DB)
	fmt.Printf("Pool Size: %d\n", c.Redis.PoolSize)
	fmt.Println()

	fmt.Println("=== Migration Settings ===")
	fmt.Printf("Auto Migrate: %v\n", c.Migrations.AutoMigrate)
	fmt.Printf("Seed Data: %v\n", c.Migrations.SeedData)
	fmt.Printf("Create Indexes: %v\n", c.Migrations.CreateIndexes)
	fmt.Println("==================================")
}
