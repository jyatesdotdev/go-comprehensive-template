package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds application configuration loaded from file, env, or flags.
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Log      LogConfig      `mapstructure:"log"`
}

// AppConfig holds general application settings.
type AppConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig holds database connection settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// LoadConfig reads configuration from file, environment variables, and defaults.
// Config file is searched in cfgFile (if set), $HOME/.myapp, and current directory.
// Environment variables are prefixed with envPrefix (e.g., MYAPP_APP_PORT).
func LoadConfig(cfgFile, envPrefix string) (*Config, error) {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME/.myapp")
	}

	// Environment variable binding: MYAPP_APP_PORT -> app.port
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Defaults
	viper.SetDefault("app.name", "myapp")
	viper.SetDefault("app.port", 8080)
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "text")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("reading config: %w", err)
		}
		// Config file not found is OK — use defaults and env vars
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}
	return &cfg, nil
}

// ConfigFileUsed returns the config file viper loaded, or empty string if none.
func ConfigFileUsed() string {
	return viper.ConfigFileUsed()
}
