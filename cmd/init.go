/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/schnauzersoft/ami-util/internal/config"

	"github.com/spf13/cobra"
)

// initCmd represents the init command.
var initCmd = &cobra.Command{
	Use:   "init [filename]",
	Short: "Initialize a configuration file",
	Long: `Initialize a configuration file with default values.

This command creates a sample configuration file that you can customize
for your environment. The file can be in YAML, YML, or TOML format.

Examples:
  ami-util init                    # Creates ami.yaml
  ami-util init my-config.yaml     # Creates my-config.yaml
  ami-util init config.toml        # Creates config.toml`,
	Args: cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		err := runInit(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(args []string) error {
	filename := "ami.yaml"
	if len(args) > 0 {
		filename = args[0]
	}

	// Create sample configuration
	sampleConfig := &config.Config{
		Accounts: []string{"137112412989"}, // Amazon Linux AMI account
		File:     "config.yaml",
		Profile:  "default",
		Verbose:  false,
		Regions:  []string{},
		RoleARN:  "",
		Patterns: []string{
			"al2023-ami-*",
			"al2023-ami-kernel-*",
			"al2023-ami-minimal-*",
			"al2023-ami-docker-*",
			"al2023-ami-ecs-*",
			"al2023-ami-eks-*",
		},
	}

	// Save configuration
	err := config.SaveConfig(sampleConfig, filename)
	if err != nil {
		return fmt.Errorf("failed to create configuration file: %w", err)
	}

	log.Printf("Configuration file created: %s", filename)
	log.Printf("Edit the file to customize your settings, then run:")
	log.Printf("  ami-util --file your-target-file.yaml")

	return nil
}
