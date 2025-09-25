# ami-util

A CLI tool that automatically updates AMI IDs (ami-xxxxx) in configuration files with the latest available versions from AWS.

[![Go](https://img.shields.io/badge/go-1.25-00ADD8.svg?logo=go)](https://tip.golang.org/doc/go1.25)
[![Go Report Card](https://goreportcard.com/badge/github.com/schnauzersoft/ami-util)](https://goreportcard.com/report/github.com/schnauzersoft/ami-util)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE.md)
[![Test Status](https://github.com/schnauzersoft/ami-util/actions/workflows/ci.yml/badge.svg)](https://github.com/schnauzersoft/ami-util/actions/workflows/ci.yml)

## Features

- **Automatic AMI Updates**: Replace old AMI IDs with the latest versions
- **Multi-Account Support**: Query multiple AWS accounts for AMI information
- **Multi-Region Support**: Search across multiple AWS regions
- **File & Directory Processing**: Update single files or entire directories
- **Flexible Authentication**: Support for AWS profiles, environment variables, and IAM roles
- **Configurable Patterns**: Customize which AMI name patterns to search for
- **Configuration Files**: YAML, TOML, and JSON configuration support
- **Safe Updates**: Creates backup files before making changes

## Installation

### Build from Source

```bash
git clone https://github.com/schnauzersoft/ami-util.git
cd ami-util
go build -o ami-util .
```

### Install Globally

```bash
go install github.com/schnauzersoft/ami-util@latest
```

## Quick Start

### 1. Initialize Configuration

```bash
# Create a default configuration file
ami-util init

# Or specify a custom filename
ami-util init my-config.yaml
```

This creates a configuration file with default settings:

```yaml
accounts:
  - "137112412989"  # Amazon Linux AMI account - add your own account IDs here
file: "config.yaml"
profile: "default"
verbose: false
regions:
  - "us-east-1"
  - "us-west-2"
role_arn: ""
patterns:
  - "al2023-ami-*"
  - "al2023-ami-kernel-*"
  - "al2023-ami-minimal-*"
  - "al2023-ami-docker-*"
  - "al2023-ami-ecs-*"
  - "al2023-ami-eks-*"
```

### 2. Update Your Configuration

Edit the generated configuration file to match your needs:

```yaml
accounts:
  - "123456789012"  # Your AWS account ID (replace with real ID)
  - "987654321098"  # Another account ID (replace with real ID)
file: "terraform/main.tf"  # Path to your config file
profile: "production"      # AWS profile to use
verbose: true              # Enable detailed output
regions:
  - "us-east-1"
  - "us-west-2"
  - "eu-west-1"
patterns:
  - "my-ami-*"            # Custom AMI patterns
  - "al2023-ami-*"        # Amazon Linux 2023
```

### 3. Run the Update

```bash
# Using configuration file
ami-util

# Or override specific settings
ami-util --file terraform/main.tf --account-ids 123456789012 --verbose
```

## Usage

### Command Line Options

```bash
ami-util [flags]

Flags:
  -a, --account-ids strings   Comma-separated list of AWS account IDs
  -f, --file string           Path to the configuration file to update
  -h, --help                  Help for ami-util
  -p, --profile string        AWS profile to use for authentication (default "default")
  -r, --regions strings       Comma-separated list of AWS regions to search (default [us-east-1,us-west-2])
      --role-arn string       Role ARN to assume (overrides AWS_ROLE_ARN env var)
      --patterns strings      Comma-separated list of AMI name patterns to search for
  -v, --verbose               Enable verbose output
```

### Environment Variables

You can use environment variables instead of command-line flags:

```bash
$ export AMI_ACCOUNTS="123456789012,987654321098"
$ export AMI_FILE="terraform/main.tf"
$ export AMI_PROFILE="production"
$ export AMI_VERBOSE="true"
$ export AMI_REGIONS="us-east-1,us-west-2,eu-west-1"
$ export AMI_ROLE_ARN="arn:aws:iam::123456789012:role/AMIAccessRole"
$ export AMI_PATTERNS="my-app-*,al2023-ami-*"

$ ami-util
```

### Configuration File Priority

Settings are loaded in the following order (later overrides earlier):

1. Default values
2. Configuration file (`ami.yaml`, `ami.yml`, `ami.toml`, `ami.json`)
3. Environment variables (`AMI_*`)
4. Command-line flags

## Examples

### Update Terraform Configuration

```bash
# Update a Terraform file
$ ami-util --file terraform/main.tf --account-ids 123456789012 --verbose

# Update with custom patterns
$ ami-util --file terraform/main.tf \
    --account-ids 123456789012 \
    --patterns "my-app-*,al2023-ami-*" \
    --regions "us-east-1,us-west-2"
```

### Update CloudFormation Template

```bash
# Update a CloudFormation template
$ ami-util --file cloudformation/template.yaml \
    --account-ids 123456789012,987654321098 \
    --profile production
```

### Update Directory of Files

```bash
# Update all files in a directory
$ ami-util --file ./configs/ \
    --account-ids 123456789012 \
    --patterns "al2023-ami-*"
```

### Using IAM Roles

```bash
# Assume a role in target accounts
$ ami-util --file config.yaml \
    --account-ids 123456789012 \
    --role-arn "arn:aws:iam::123456789012:role/AMIAccessRole"
```

## AWS Authentication

The tool supports multiple authentication methods:

### 1. AWS Profiles
```bash
$ ami-util --profile production
```

### 2. Environment Variables
```bash
$ export AWS_PROFILE="production"
$ export AWS_ACCESS_KEY_ID="your-key"
$ export AWS_SECRET_ACCESS_KEY="your-secret"
$ export AWS_SESSION_TOKEN="your-token"  # For temporary credentials
```

### 3. IAM Roles
```bash
# Using AWS_ROLE_ARN environment variable
$ export AWS_ROLE_ARN="arn:aws:iam::123456789012:role/AMIAccessRole"
$ ami-util

# Or via command line
$ ami-util --role-arn "arn:aws:iam::123456789012:role/AMIAccessRole"
```

### 4. EC2 Instance Profile
If running on an EC2 instance with an IAM role attached, no additional configuration is needed.

## Required IAM Permissions

The tool needs the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeImages"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "sts:AssumeRole"
      ],
      "Resource": "arn:aws:iam::*:role/AMIAccessRole"
    }
  ]
}
```

## Configuration File Formats

### YAML (ami.yaml)
```yaml
accounts: ["123456789012"]
file: "config.yaml"
profile: "default"
verbose: false
regions: ["us-east-1", "us-west-2"]
role_arn: ""
patterns: ["al2023-ami-*"]
```

### TOML (ami.toml)
```toml
accounts = ["123456789012"]
file = "config.yaml"
profile = "default"
verbose = false
regions = ["us-east-1", "us-west-2"]
role_arn = ""
patterns = ["al2023-ami-*"]
```

### JSON (ami.json)
```json
{
  "accounts": ["123456789012"],
  "file": "config.yaml",
  "profile": "default",
  "verbose": false,
  "regions": ["us-east-1", "us-west-2"],
  "role_arn": "",
  "patterns": ["al2023-ami-*"]
}
```
## Troubleshooting

### Common Issues

**"No AMI replacements found"**
- Check that your patterns match existing AMI names
- Verify AWS permissions for `ec2:DescribeImages`
- Ensure the account IDs are correct

**"Failed to assume role"**
- Verify the role ARN is correct
- Check that your current credentials can assume the role
- Ensure the role has the necessary trust policy

**"File path does not exist"**
- Verify the file path is correct
- Check file permissions
- Ensure the file is readable

### Debug Mode

Use `--verbose` to see detailed information about the process:

```bash
$ ami-util --file config.yaml --verbose
```

This will show:
- Configuration being used
- AWS accounts and regions being queried
- AMI patterns being searched
- Number of replacements made per file
