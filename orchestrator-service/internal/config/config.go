package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	GRPC     GRPCConfig     `mapstructure:"grpc"`
	Security SecurityConfig `mapstructure:"security"`
	Logging  LoggingConfig  `mapstructure:"logging"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type GRPCConfig struct {
	Host    string `mapstructure:"host"`
	Port    int    `mapstructure:"port"`
	HostKey string `mapstructure:"host_key"`
	CertKey string `mapstructure:"cert_key"`
}

type SecurityConfig struct {
	JWTSecret     string `mapstructure:"jwt_secret"`
	NodeAuthToken string `mapstructure:"node_auth_token"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json, text
	Output string `mapstructure:"output"` // stdout, file
}

func LoadConfig() (*Config, error) {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		// .env file is optional
		fmt.Println("Warning: .env file not found")
	}

	// Set default values
	setDefaults()

	// Load config from file and environment
	viper.SetConfigName("orchestrator")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath("../configs")
	viper.AddConfigPath("../../configs")

	// Enable environment variable binding
	viper.AutomaticEnv()

	// Bind environment variables
	bindEnvVars()

	// Read config file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("Config file not found, using environment variables and defaults")
		} else {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("server.mode", "release")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "hysteryvpn")
	viper.SetDefault("database.sslmode", "disable")

	viper.SetDefault("grpc.host", "0.0.0.0")
	viper.SetDefault("grpc.port", 50052)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")
}

func bindEnvVars() {
	viper.BindEnv("server.host", "SERVER_HOST")
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.mode", "GIN_MODE")

	viper.BindEnv("database.host", "DB_HOST")
	viper.BindEnv("database.port", "DB_PORT")
	viper.BindEnv("database.user", "DB_USER")
	viper.BindEnv("database.password", "DB_PASSWORD")
	viper.BindEnv("database.dbname", "DB_NAME")
	viper.BindEnv("database.sslmode", "DB_SSLMODE")

	viper.BindEnv("grpc.host", "GRPC_HOST")
	viper.BindEnv("grpc.port", "GRPC_PORT")
	viper.BindEnv("grpc.host_key", "GRPC_HOST_KEY")
	viper.BindEnv("grpc.cert_key", "GRPC_CERT_KEY")

	viper.BindEnv("security.jwt_secret", "JWT_SECRET")
	viper.BindEnv("security.node_auth_token", "NODE_AUTH_TOKEN")

	viper.BindEnv("logging.level", "LOG_LEVEL")
	viper.BindEnv("logging.format", "LOG_FORMAT")
	viper.BindEnv("logging.output", "LOG_OUTPUT")
}

// Helper functions for environment variables
func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
