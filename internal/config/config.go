/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	DefaultDirPerm = 0o755
)

var (
	ErrNoAccountID = errors.New("at least one account ID is required")
	ErrNoFilePath  = errors.New("file path is required")
)

type Config struct {
	Accounts []string `mapstructure:"accounts" toml:"accounts" yaml:"accounts"`
	File     string   `mapstructure:"file"     toml:"file"     yaml:"file"`
	Profile  string   `mapstructure:"profile"  toml:"profile"  yaml:"profile"`
	Verbose  bool     `mapstructure:"verbose"  toml:"verbose"  yaml:"verbose"`
	Regions  []string `mapstructure:"regions"  toml:"regions"  yaml:"regions"`
	RoleARN  string   `mapstructure:"role_arn" toml:"role_arn" yaml:"roleArn"`
	Patterns []string `mapstructure:"patterns" toml:"patterns" yaml:"patterns"`
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

	_ = viper.BindEnv("accounts", "AMI_ACCOUNTS")
	_ = viper.BindEnv("file", "AMI_FILE")
	_ = viper.BindEnv("profile", "AMI_PROFILE")
	_ = viper.BindEnv("verbose", "AMI_VERBOSE")
	_ = viper.BindEnv("regions", "AMI_REGIONS")
	_ = viper.BindEnv("role_arn", "AMI_ROLE_ARN")
	_ = viper.BindEnv("patterns", "AMI_PATTERNS")

	err := viper.ReadInConfig()
	if err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if !errors.As(err, &configFileNotFoundError) {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func SaveConfig(config *Config, filename string) error {
	dir := filepath.Dir(filename)

	err := os.MkdirAll(dir, DefaultDirPerm)
	if err != nil {
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

	err = viper.WriteConfigAs(filename)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func ValidateConfig(config *Config) error {
	if len(config.Accounts) == 0 {
		return ErrNoAccountID
	}

	if config.File == "" {
		return ErrNoFilePath
	}

	return nil
}
