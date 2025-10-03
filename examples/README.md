# ami-util Examples

This directory contains example configuration files and demonstrates how to use `ami-util` to update AMI IDs in your configuration files.

## Overview

`ami-util` is a CLI tool that automatically updates AMI IDs in configuration files to their latest versions. It supports both YAML and TOML configuration formats and can work with AWS profiles or assume roles.

## Example Files

- `ami.yaml` - Example configuration file in YAML format
- `config.yaml` - Example target file containing AMI IDs to update

## Basic Usage

### 1. Using Configuration File

The simplest way to use `ami-util` is with a configuration file:

```bash
$ ami-util
```

### 2. Using Command Line Flags

You can override configuration file settings with command line flags:

```bash
$ ami-util --file config.yaml --account-ids 092701018921 --regions us-east-1 --verbose
```

### 3. Using Environment Variables

Set environment variables to override configuration:

```bash
$ export AMI_ACCOUNTS="092701018921"
$ export AMI_FILE="config.yaml"
$ export AMI_REGIONS="us-east-1,us-west-2"
$ export AMI_PROFILE="dev"
$ ami-util
```

## Example Workflows

### Example 1: Update AMI IDs in a Single File

**Before running:**
```yaml
image_id: "ami-037057f9512b47316"
```

**Command:**
```bash
$ ami-util --file config.yaml --account-ids 092701018921 --regions us-east-1 --verbose
```

**Expected output:**
```
2025/10/03 13:20:45 Updating AMI IDs in file: config.yaml
2025/10/03 13:20:45 Account IDs: 092701018921
2025/10/03 13:20:45 Regions: us-east-1
2025/10/03 13:20:45 AWS Profile: default
2025/10/03 13:20:45 Processing account: 092701018921
2025/10/03 13:20:45   Processing region: us-east-1
2025/10/03 13:20:46     Found 1 AMI replacements
2025/10/03 13:20:46 Updated 1 AMI references in config.yaml (backup created at config.yaml.backup)
2025/10/03 13:20:46 Successfully processed config.yaml
```

**After running:**
```yaml
image_id: "ami-0ea3a93c835afbde0"
```

### Example 2: Using Configuration File

**Configuration file (ami.yaml):**
```yaml
accounts:
    - "092701018921"
file: config.yaml
patterns:
    - al2023-ami-*
    - bottlerocket-aws-ecs-2-aarch64-*
profile: dev
regions:
    - us-east-1
    - us-west-2
role_arn: ""
verbose: true
```

**Command:**
```bash
$ ami-util
```

**Expected output:**
```
2025/10/03 13:20:45 Updating AMI IDs in file: config.yaml
2025/10/03 13:20:45 Account IDs: 092701018921
2025/10/03 13:20:45 Regions: us-east-1, us-west-2
2025/10/03 13:20:45 AWS Profile: dev
2025/10/03 13:20:45 Processing account: 092701018921
2025/10/03 13:20:45   Processing region: us-east-1
2025/10/03 13:20:45     Found 1 AMI replacements
2025/10/03 13:20:45   Processing region: us-west-2
2025/10/03 13:20:46     Found 0 AMI replacements
2025/10/03 13:20:46 Updated 1 AMI references in config.yaml (backup created at config.yaml.backup)
2025/10/03 13:20:46 Successfully processed config.yaml
```

### Example 3: Using AWS Profile Region

If you don't specify regions, the tool will use the region from your AWS profile:

**Command:**
```bash
$ ami-util --file config.yaml --account-ids 092701018921
```

**Expected output:**
```
2025/10/03 13:20:45 Updating AMI IDs in file: config.yaml
2025/10/03 13:20:45 Account IDs: 092701018921
2025/10/03 13:20:45 Regions: will use region from AWS profile
2025/10/03 13:20:45 AWS Profile: default
2025/10/03 13:20:45 Processing account: 092701018921
2025/10/03 13:20:45   Processing region: us-east-1
2025/10/03 13:20:46     Found 1 AMI replacements
2025/10/03 13:20:46 Updated 1 AMI references in config.yaml (backup created at config.yaml.backup)
2025/10/03 13:20:46 Successfully processed config.yaml
```

### Example 4: Using IAM Roles

**Command:**
```bash
$ ami-util --file config.yaml --account-ids 123456789012 --role-arn "arn:aws:iam::123456789012:role/AMIAccessRole"
```

**Expected output:**
```
2025/10/03 13:20:45 Updating AMI IDs in file: config.yaml
2025/10/03 13:20:45 Account IDs: 123456789012
2025/10/03 13:20:45 Regions: will use region from AWS profile
2025/10/03 13:20:45 AWS Profile: default
2025/10/03 13:20:45 Role ARN: arn:aws:iam::123456789012:role/AMIAccessRole
2025/10/03 13:20:45 Processing account: 123456789012
2025/10/03 13:20:45   Processing region: us-east-1
2025/10/03 13:20:46     Found 1 AMI replacements
2025/10/03 13:20:46 Updated 1 AMI references in config.yaml (backup created at config.yaml.backup)
2025/10/03 13:20:46 Successfully processed config.yaml
```

## Configuration File Formats

### YAML Format (ami.yaml)
```yaml
accounts:
    - "092701018921"
    - "123456789012"
file: config.yaml
patterns:
    - al2023-ami-*
    - bottlerocket-aws-ecs-2-aarch64-*
profile: dev
regions:
    - us-east-1
    - us-west-2
role_arn: ""
verbose: true
```

### TOML Format (ami.toml)
```toml
accounts = ["092701018921", "123456789012"]
file = "config.yaml"
patterns = [
    "al2023-ami-*",
    "bottlerocket-aws-ecs-2-aarch64-*"
]
profile = "dev"
regions = ["us-east-1", "us-west-2"]
role_arn = ""
verbose = true
```

## AMI Pattern Matching

The tool supports two types of pattern matching:

### 1. AMI ID Patterns
When the tool finds an actual AMI ID (like `ami-037057f9512b47316`), it:
1. Looks up the AMI to get its name
2. Finds newer versions of the same AMI type
3. Replaces the old AMI with the latest version

### 2. Name Patterns
When using patterns like `al2023-ami-*`, the tool:
1. Searches for all AMIs matching the pattern
2. Sorts them by creation date
3. Replaces older AMIs with the latest one

## Backup Files

The tool automatically creates backup files before making changes:
- Original file: `config.yaml`
- Backup file: `config.yaml.backup`

## Error Handling

The tool handles various error conditions gracefully:
- AMIs not found in specific regions
- Invalid AWS credentials
- Network connectivity issues
- File permission errors

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `AMI_ACCOUNTS` | Comma-separated list of AWS account IDs | `"092701018921,123456789012"` |
| `AMI_FILE` | Path to the file to update | `"config.yaml"` |
| `AMI_PROFILE` | AWS profile to use | `"dev"` |
| `AMI_REGIONS` | Comma-separated list of regions | `"us-east-1,us-west-2"` |
| `AMI_ROLE_ARN` | Role ARN to assume | `"arn:aws:iam::123456789012:role/AMIAccessRole"` |
| `AMI_PATTERNS` | Comma-separated list of patterns | `"al2023-ami-*,bottlerocket-*"` |
| `AMI_VERBOSE` | Enable verbose output | `"true"` |
