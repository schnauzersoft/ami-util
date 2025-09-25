/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Accounts []string `mapstructure:"accounts" yaml:"accounts" toml:"accounts"`
	File     string   `mapstructure:"file" yaml:"file" toml:"file"`
	Profile  string   `mapstructure:"profile" yaml:"profile" toml:"profile"`
	Verbose  bool     `mapstructure:"verbose" yaml:"verbose" toml:"verbose"`
	Regions  []string `mapstructure:"regions" yaml:"regions" toml:"regions"`
	RoleARN  string   `mapstructure:"role_arn" yaml:"role_arn" toml:"role_arn"`
	Patterns []string `mapstructure:"patterns" yaml:"patterns" toml:"patterns"`
}

func LoadConfig() (*Config, error) {
	viper.SetDefault("profile", "default")
	viper.SetDefault("verbose", false)
	viper.SetDefault("regions", []string{"us-east-1", "us-west-2"})
	viper.SetDefault("patterns", []string{
		"al2023-ami-*",
		"al2023-ami-kernel-*",
		"al2023-ami-minimal-*",
		"al2023-ami-docker-*",
		"al2023-ami-ecs-*",
		"al2023-ami-eks-*",
	})

	viper.SetConfigName("ami")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.ami-util")
	viper.AddConfigPath("/etc/ami-util")

	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.ami-util")
	viper.AddConfigPath("/etc/ami-util")

	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.ami-util")
	viper.AddConfigPath("/etc/ami-util")

	viper.SetEnvPrefix("AMI")
	viper.AutomaticEnv()

	viper.BindEnv("accounts", "AMI_ACCOUNTS")
	viper.BindEnv("file", "AMI_FILE")
	viper.BindEnv("profile", "AMI_PROFILE")
	viper.BindEnv("verbose", "AMI_VERBOSE")
	viper.BindEnv("regions", "AMI_REGIONS")
	viper.BindEnv("role_arn", "AMI_ROLE_ARN")
	viper.BindEnv("patterns", "AMI_PATTERNS")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func SaveConfig(config *Config, filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.SetConfigFile(filename)

	viper.Set("accounts", config.Accounts)
	viper.Set("file", config.File)
	viper.Set("profile", config.Profile)
	viper.Set("verbose", config.Verbose)
	viper.Set("regions", config.Regions)
	viper.Set("role_arn", config.RoleARN)
	viper.Set("patterns", config.Patterns)

	if err := viper.WriteConfigAs(filename); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func ValidateConfig(config *Config) error {
	if len(config.Accounts) == 0 {
		return fmt.Errorf("at least one account ID is required")
	}
	if config.File == "" {
		return fmt.Errorf("file path is required")
	}
	return nil
}
