/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/schnauzersoft/ami-util/internal/aws"
	"github.com/schnauzersoft/ami-util/internal/config"
	"github.com/schnauzersoft/ami-util/internal/fileprocessor"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfg *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ami-util",
	Short: "Update AMI IDs to latest versions in configuration files",
	Long: `ami-util is a CLI tool that replaces AMI IDs (ami-xxxxx) in configuration files
with the latest available versions from AWS.

The tool assumes roles in specified AWS accounts, discovers the latest AMI images
matching specific patterns, and updates configuration files accordingly.

Configuration:
  The tool supports multiple configuration sources (in order of precedence):
  1. Command line flags
  2. Environment variables (AMI_* prefix)
  3. Configuration file (ami.yaml, ami.yml, or ami.toml)
  4. Default values

  Configuration file locations (searched in order):
  - ./ami.yaml (or .yml, .toml)
  - $HOME/.ami-util/ami.yaml
  - /etc/ami-util/ami.yaml

AWS Authentication:
  The tool uses standard AWS environment variables and configuration:
  - AWS_PROFILE: AWS profile to use (default: "default")
  - AWS_ROLE_ARN: Role ARN to assume (if not set, constructs from account ID)
  - AWS_ROLE_SESSION_NAME: Session name for role assumption (default: "UpdateToLatestAMI")
  - AWS_ROLE_EXTERNAL_ID: External ID for role assumption (optional)

Examples:
  # Using command line flags
  ami-util --account-ids 123456789012,987654321098 --file config.yaml
  
  # Using configuration file
  ami-util --file config.yaml  # account-ids from ami.yaml
  
  # Using environment variables
  AMI_ACCOUNTS=123456789012,987654321098 ami-util --file config.yaml
  
  # Mixed usage
  ami-util --account-ids 123456789012 --file config.yaml --profile myprofile`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runUpdate(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Initialize Viper
	viper.SetConfigName("ami")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.ami-util")
	viper.AddConfigPath("/etc/ami-util")

	// Set environment variable prefix
	viper.SetEnvPrefix("AMI")
	viper.AutomaticEnv()

	// Bind environment variables
	viper.BindEnv("accounts", "AMI_ACCOUNTS")
	viper.BindEnv("file", "AMI_FILE")
	viper.BindEnv("profile", "AMI_PROFILE")
	viper.BindEnv("verbose", "AMI_VERBOSE")
	viper.BindEnv("regions", "AMI_REGIONS")
	viper.BindEnv("role_arn", "AMI_ROLE_ARN")

	// Set default values
	viper.SetDefault("profile", "default")
	viper.SetDefault("verbose", false)
	viper.SetDefault("regions", []string{"us-east-1", "us-west-2"})

	// Define flags
	rootCmd.Flags().StringSlice("account-ids", []string{}, "Comma-separated list of AWS account IDs")
	rootCmd.Flags().String("file", "", "Path to the configuration file to update")
	rootCmd.Flags().String("profile", "default", "AWS profile to use for authentication")
	rootCmd.Flags().Bool("verbose", false, "Enable verbose output")
	rootCmd.Flags().StringSlice("regions", []string{"us-east-1", "us-west-2"}, "Comma-separated list of AWS regions to search")
	rootCmd.Flags().String("role-arn", "", "Role ARN to assume (overrides AWS_ROLE_ARN env var)")

	// Bind flags to viper
	viper.BindPFlag("accounts", rootCmd.Flags().Lookup("account-ids"))
	viper.BindPFlag("file", rootCmd.Flags().Lookup("file"))
	viper.BindPFlag("profile", rootCmd.Flags().Lookup("profile"))
	viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose"))
	viper.BindPFlag("regions", rootCmd.Flags().Lookup("regions"))
	viper.BindPFlag("role_arn", rootCmd.Flags().Lookup("role-arn"))

	// Mark required flags
	rootCmd.MarkFlagRequired("file")
}

func runUpdate() error {
	// Load and validate configuration
	if err := loadAndValidateConfig(); err != nil {
		return err
	}

	// Print configuration info if verbose
	printConfigInfo()

	// Create AWS client and file processor
	awsClient, fileProcessor, err := createClients()
	if err != nil {
		return err
	}

	// Get file info and patterns
	fileInfo, patterns, err := getFileInfoAndPatterns(fileProcessor)
	if err != nil {
		return err
	}

	// Collect AMI replacements from all accounts and regions
	allReplacements, err := collectAMIReplacements(awsClient, patterns)
	if err != nil {
		return err
	}

	if len(allReplacements) == 0 {
		fmt.Println("No AMI replacements found")
		return nil
	}

	// Process the file or directory
	if err := processFiles(fileProcessor, fileInfo, allReplacements); err != nil {
		return err
	}

	fmt.Printf("Successfully processed %s\n", cfg.File)
	return nil
}

func loadAndValidateConfig() error {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := config.ValidateConfig(cfg); err != nil {
		return err
	}

	return nil
}

func printConfigInfo() {
	if cfg.Verbose {
		fmt.Printf("Updating AMI IDs in file: %s\n", cfg.File)
		fmt.Printf("Account IDs: %s\n", strings.Join(cfg.Accounts, ", "))
		fmt.Printf("Regions: %s\n", strings.Join(cfg.Regions, ", "))
		fmt.Printf("AWS Profile: %s\n", cfg.Profile)
		if cfg.RoleARN != "" {
			fmt.Printf("Role ARN: %s\n", cfg.RoleARN)
		}
	}
}

func createClients() (*aws.Client, *fileprocessor.Processor, error) {
	awsClient, err := aws.NewClient(cfg.Profile, cfg.RoleARN)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create AWS client: %w", err)
	}

	fileProcessor := fileprocessor.NewProcessor(cfg.Verbose)
	return awsClient, fileProcessor, nil
}

func getFileInfoAndPatterns(fileProcessor *fileprocessor.Processor) (os.FileInfo, []string, error) {
	fileInfo, err := os.Stat(cfg.File)
	if err != nil {
		return nil, nil, fmt.Errorf("file path does not exist: %w", err)
	}

	var patterns []string
	if !fileInfo.IsDir() {
		// Extract AMI patterns from the file
		filePatterns, err := fileProcessor.FindAMIsInFile(cfg.File)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to find AMIs in file: %w", err)
		}
		patterns = filePatterns
	} else {
		// Use default patterns for directory processing
		patterns = aws.GenerateAMIPatterns()
	}

	return fileInfo, patterns, nil
}

func collectAMIReplacements(awsClient *aws.Client, patterns []string) ([]aws.AMIReplacement, error) {
	var allReplacements []aws.AMIReplacement

	for _, accountID := range cfg.Accounts {
		if cfg.Verbose {
			fmt.Printf("Processing account: %s\n", accountID)
		}

		accountReplacements, err := processAccount(awsClient, accountID, patterns)
		if err != nil {
			return nil, err
		}

		allReplacements = append(allReplacements, accountReplacements...)
	}

	return allReplacements, nil
}

func processAccount(awsClient *aws.Client, accountID string, patterns []string) ([]aws.AMIReplacement, error) {
	var accountReplacements []aws.AMIReplacement

	for _, region := range cfg.Regions {
		if cfg.Verbose {
			fmt.Printf("  Processing region: %s\n", region)
		}

		replacements, err := awsClient.GetLatestAMIs(accountID, region, patterns)
		if err != nil {
			fmt.Printf("Warning: failed to get AMIs for account %s, region %s: %v\n", accountID, region, err)
			continue
		}

		accountReplacements = append(accountReplacements, replacements...)

		if cfg.Verbose {
			fmt.Printf("    Found %d AMI replacements\n", len(replacements))
		}
	}

	return accountReplacements, nil
}

func processFiles(fileProcessor *fileprocessor.Processor, fileInfo os.FileInfo, allReplacements []aws.AMIReplacement) error {
	var err error
	if fileInfo.IsDir() {
		err = fileProcessor.ProcessDirectory(cfg.File, allReplacements)
	} else {
		err = fileProcessor.ProcessFile(cfg.File, allReplacements)
	}

	if err != nil {
		return fmt.Errorf("failed to process file: %w", err)
	}

	return nil
}
