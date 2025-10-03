/*
Copyright © 2025 Ben Sapp ya.bsapp.ru
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/schnauzersoft/ami-util/internal/aws"
	"github.com/schnauzersoft/ami-util/internal/config"
	"github.com/schnauzersoft/ami-util/internal/fileprocessor"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildTime = "unknown"
)

var cfg *config.Config

// rootCmd represents the base command when called without any subcommands.
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

AMI Patterns:
  The tool searches for AMI images matching specified patterns. Default patterns
  include Amazon Linux 2023 AMI types. You can override these with:
  - --patterns flag: Comma-separated list of patterns
  - AMI_PATTERNS environment variable
  - patterns key in configuration file

Examples:
  # Using command line flags
  ami-util --account-ids 123456789012,987654321098 --file config.yaml

  # Using custom patterns
  ami-util --account-ids 123456789012 --file config.yaml --patterns "my-app-*","my-service-*"
  
  # Using configuration file
  ami-util --file config.yaml  # account-ids and patterns from ami.yaml
  
  # Using environment variables
  AMI_ACCOUNTS=123456789012,987654321098 AMI_PATTERNS="my-app-*" ami-util --file config.yaml
  
  # Mixed usage
  ami-util --account-ids 123456789012 --file config.yaml --profile myprofile`,
	Run: func(_ *cobra.Command, _ []string) {
		err := runUpdate()
		if err != nil {
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
	_ = viper.BindEnv("accounts", "AMI_ACCOUNTS")
	_ = viper.BindEnv("file", "AMI_FILE")
	_ = viper.BindEnv("profile", "AMI_PROFILE")
	_ = viper.BindEnv("verbose", "AMI_VERBOSE")
	_ = viper.BindEnv("regions", "AMI_REGIONS")
	_ = viper.BindEnv("role_arn", "AMI_ROLE_ARN")
	_ = viper.BindEnv("patterns", "AMI_PATTERNS")

	// Set default values
	viper.SetDefault("profile", "default")
	viper.SetDefault("verbose", false)

	// Define flags
	rootCmd.Flags().StringSlice("account-ids", []string{}, "Comma-separated list of AWS account IDs")
	rootCmd.Flags().String("file", "", "Path to the configuration file to update")
	rootCmd.Flags().String("profile", "default", "AWS profile to use for authentication")
	rootCmd.Flags().Bool("verbose", false, "Enable verbose output")
	rootCmd.Flags().StringSlice("regions", []string{},
		"Comma-separated list of AWS regions to search (if not specified, will use region from AWS profile)")
	rootCmd.Flags().String("role-arn", "", "Role ARN to assume (overrides AWS_ROLE_ARN env var)")
	rootCmd.Flags().StringSlice("patterns", []string{}, "Comma-separated list of AMI name patterns to search for")

	// Bind flags to viper
	_ = viper.BindPFlag("accounts", rootCmd.Flags().Lookup("account-ids"))
	_ = viper.BindPFlag("file", rootCmd.Flags().Lookup("file"))
	_ = viper.BindPFlag("profile", rootCmd.Flags().Lookup("profile"))
	_ = viper.BindPFlag("verbose", rootCmd.Flags().Lookup("verbose"))
	_ = viper.BindPFlag("regions", rootCmd.Flags().Lookup("regions"))
	_ = viper.BindPFlag("role_arn", rootCmd.Flags().Lookup("role-arn"))
	_ = viper.BindPFlag("patterns", rootCmd.Flags().Lookup("patterns"))

	// Mark required flags
	_ = rootCmd.MarkFlagRequired("file")
}

func runUpdate() error {
	// Load and validate configuration
	err := loadAndValidateConfig()
	if err != nil {
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
	allReplacements := collectAMIReplacements(awsClient, patterns)

	if len(allReplacements) == 0 {
		log.Println("No AMI replacements found")

		return nil
	}

	// Process the file or directory
	err = processFiles(fileProcessor, fileInfo, allReplacements)
	if err != nil {
		return err
	}

	log.Printf("Successfully processed %s", cfg.File)

	return nil
}

func loadAndValidateConfig() error {
	var err error

	cfg, err = config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	err = config.ValidateConfig(cfg)
	if err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	return nil
}

func printConfigInfo() {
	if cfg.Verbose {
		log.Printf("Updating AMI IDs in file: %s", cfg.File)
		log.Printf("Account IDs: %s", strings.Join(cfg.Accounts, ", "))

		if len(cfg.Regions) > 0 {
			log.Printf("Regions: %s", strings.Join(cfg.Regions, ", "))
		} else {
			log.Printf("Regions: will use region from AWS profile")
		}

		log.Printf("AWS Profile: %s", cfg.Profile)

		if cfg.RoleARN != "" {
			log.Printf("Role ARN: %s", cfg.RoleARN)
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
		// Use configured patterns for directory processing
		patterns = cfg.Patterns
	}

	return fileInfo, patterns, nil
}

func collectAMIReplacements(awsClient *aws.Client, patterns []string) []aws.AMIReplacement {
	var allReplacements []aws.AMIReplacement

	for _, accountID := range cfg.Accounts {
		if cfg.Verbose {
			log.Printf("Processing account: %s", accountID)
		}

		accountReplacements := processAccount(awsClient, accountID, patterns)
		allReplacements = append(allReplacements, accountReplacements...)
	}

	return allReplacements
}

func processAccount(awsClient *aws.Client, accountID string, patterns []string) []aws.AMIReplacement {
	var accountReplacements []aws.AMIReplacement

	regions := cfg.Regions
	if len(regions) == 0 {
		// No regions specified, get region from AWS profile
		region, err := awsClient.GetRegion()
		if err != nil {
			log.Printf("Warning: failed to get region from AWS profile: %v", err)

			return accountReplacements
		}

		regions = []string{region}
	}

	for _, region := range regions {
		if cfg.Verbose {
			log.Printf("  Processing region: %s", region)
		}

		replacements, err := awsClient.GetLatestAMIs(accountID, region, patterns)
		if err != nil {
			log.Printf("Warning: failed to get AMIs for account %s, region %s: %v", accountID, region, err)

			continue
		}

		accountReplacements = append(accountReplacements, replacements...)

		if cfg.Verbose {
			log.Printf("    Found %d AMI replacements", len(replacements))
		}
	}

	return accountReplacements
}

func processFiles(fileProcessor *fileprocessor.Processor, fileInfo os.FileInfo,
	allReplacements []aws.AMIReplacement,
) error {
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
