package config

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Security    SecurityConfig
	RootEmail   string
	Environment string
}

type ServerConfig struct {
	Port int
	Host string
	SSL  SSLConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Prefix   string
}

type SecurityConfig struct {
	// Raw decoded bytes for PASETO keys
	PasetoPrivateKey []byte
	PasetoPublicKey  []byte
}

type SSLConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

// LoadOptions contains options for loading configuration
type LoadOptions struct {
	EnvFile string // Optional environment file to load (e.g., ".env", ".env.test")
}

// Load loads the configuration with default options
func Load() (*Config, error) {
	return LoadWithOptions(LoadOptions{EnvFile: ".env"})
}

// LoadWithOptions loads the configuration with the specified options
func LoadWithOptions(opts LoadOptions) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("SERVER_HOST", "0.0.0.0")
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 5432)
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_PASSWORD", "postgres")
	v.SetDefault("DB_PREFIX", "notifuse")
	v.SetDefault("DB_NAME", "${DB_PREFIX}_system")
	v.SetDefault("ENVIRONMENT", "production")

	// Load environment file if specified
	if opts.EnvFile != "" {
		v.SetConfigName(opts.EnvFile)
		v.SetConfigType("env")

		currentPath, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("error getting current directory: %w", err)
		}

		configPath := strings.Split(currentPath, "server")[0] + "server/"
		v.AddConfigPath(configPath)

		if err := v.ReadInConfig(); err != nil {
			log.Printf("Error reading config file %v: %v", opts.EnvFile, err)
			// It's okay if config file doesn't exist
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	// Read environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Get base64 encoded keys
	privateKeyBase64 := v.GetString("PASETO_PRIVATE_KEY")
	publicKeyBase64 := v.GetString("PASETO_PUBLIC_KEY")

	// Validate required configuration
	if privateKeyBase64 == "" {
		return nil, fmt.Errorf("PASETO_PRIVATE_KEY is required")
	}
	if publicKeyBase64 == "" {
		return nil, fmt.Errorf("PASETO_PUBLIC_KEY is required")
	}

	// Decode base64 keys
	privateKey, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("error decoding PASETO_PRIVATE_KEY: %w", err)
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("error decoding PASETO_PUBLIC_KEY: %w", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port: v.GetInt("SERVER_PORT"),
			Host: v.GetString("SERVER_HOST"),
			SSL: SSLConfig{
				Enabled:  v.GetBool("SSL_ENABLED"),
				CertFile: v.GetString("SSL_CERT_FILE"),
				KeyFile:  v.GetString("SSL_KEY_FILE"),
			},
		},
		Database: DatabaseConfig{
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetInt("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			DBName:   v.GetString("DB_NAME"),
			Prefix:   v.GetString("DB_PREFIX"),
		},
		Security: SecurityConfig{
			PasetoPrivateKey: privateKey,
			PasetoPublicKey:  publicKey,
		},
		RootEmail:   v.GetString("ROOT_EMAIL"),
		Environment: v.GetString("ENVIRONMENT"),
	}

	return config, nil
}

// IsDevelopment returns true if the environment is set to development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}
