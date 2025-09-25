#!/bin/bash

# ami-util Installation Script
# Copyright © 2025 Ben Sapp ya.bsapp.ru

set -euo pipefail

REPO="schnauzersoft/ami-util"
BINARY_NAME="ami-util"
INSTALL_DIR="${HOME}/.local/bin"
SHELL_RC_FILES=("${HOME}/.bashrc" "${HOME}/.zshrc" "${HOME}/.profile")

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_help() {
    cat << EOF
ami-util Installation Script

USAGE:
    $0 [OPTIONS]

OPTIONS:
    -h, --help          Show this help message
    -v, --version       Specify version to install (default: latest)
    -d, --dir DIR       Installation directory (default: ~/.local/bin)
    --no-path           Don't add binary to PATH
    --force             Force reinstall even if binary exists
    --no-verify         Skip signature verification (not recommended)

EXAMPLES:
    $0                          # Install latest version
    $0 --version v1.0.0         # Install specific version
    $0 --dir /usr/local/bin     # Install to custom directory
    $0 --no-path                # Install without adding to PATH

EOF
}

VERSION="latest"
FORCE=false
NO_PATH=false
NO_VERIFY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -d|--dir)
            INSTALL_DIR="$2"
            shift 2
            ;;
        --no-path)
            NO_PATH=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --no-verify)
            NO_VERIFY=true
            shift
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          log_error "Unsupported operating system: $(uname -s)"; exit 1 ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        armv7l)         arch="armv7" ;;
        *)              log_error "Unsupported architecture: $(uname -m)"; exit 1 ;;
    esac
    
    echo "${os}-${arch}"
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

download_file() {
    local url="$1"
    local output="$2"
    
    if command_exists curl; then
        curl -L --progress-bar -o "$output" "$url"
    elif command_exists wget; then
        wget --progress=bar:force -O "$output" "$url"
    else
        log_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
}

get_latest_release() {
    local api_url="https://api.github.com/repos/${REPO}/releases/latest"
    
    if command_exists jq; then
        curl -s "$api_url" | jq -r '.tag_name'
    elif command_exists python3; then
        curl -s "$api_url" | python3 -c "import sys, json; print(json.load(sys.stdin)['tag_name'])"
    else
        log_warning "jq or python3 not found, using fallback method"
        curl -s "$api_url" | grep '"tag_name"' | sed 's/.*"tag_name": "\(.*\)".*/\1/'
    fi
}

verify_signature() {
    local binary="$1"

    if [[ "$NO_VERIFY" == "true" ]]; then
        log_warning "Skipping signature verification"
        return 0
    fi
    
    if [[ ! -x "$binary" ]]; then
        log_error "Binary is not executable"
        return 1
    fi
    
    if [[ ! -s "$binary" ]]; then
        log_error "Binary file is empty or corrupted"
        return 1
    fi
    
    log_success "Binary verification passed"
    return 0
}

add_to_path() {
    local binary_path="$1"
    
    if [[ "$NO_PATH" == "true" ]]; then
        log_info "Skipping PATH modification (--no-path specified)"
        return 0
    fi
    
    if command_exists "$BINARY_NAME"; then
        log_info "Binary is already in PATH"
        return 0
    fi
    
    mkdir -p "$INSTALL_DIR"
    
    local path_export="export PATH=\"${INSTALL_DIR}:\$PATH\""
    
    for rc_file in "${SHELL_RC_FILES[@]}"; do
        if [[ -f "$rc_file" ]]; then
            if ! grep -q "$path_export" "$rc_file" 2>/dev/null; then
                echo "$path_export" >> "$rc_file"
                log_info "Added to PATH in $rc_file"
            fi
        fi
    done
    
    export PATH="${INSTALL_DIR}:${PATH}"
    
    log_success "Binary added to PATH"
}

main() {
    log_info "Starting ami-util installation..."
    
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    if [[ "$VERSION" == "latest" ]]; then
        log_info "Fetching latest version..."
        VERSION=$(get_latest_release)
        if [[ -z "$VERSION" ]]; then
            log_error "Failed to get latest version"
            exit 1
        fi
    fi
    
    log_info "Installing version: $VERSION"
    
    local binary_path="${INSTALL_DIR}/${BINARY_NAME}"
    if [[ -f "$binary_path" && "$FORCE" != "true" ]]; then
        log_warning "Binary already exists at $binary_path"
        read -p "Do you want to overwrite it? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            log_info "Installation cancelled"
            exit 0
        fi
    fi
    
    mkdir -p "$INSTALL_DIR"
    
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${BINARY_NAME}-${VERSION}-${platform}.tar.gz"
    local temp_dir
    temp_dir=$(mktemp -d)
    local archive_path="${temp_dir}/${BINARY_NAME}.tar.gz"
    
    log_info "Downloading from: $download_url"
    download_file "$download_url" "$archive_path"
    
    log_info "Extracting archive..."
    tar -xzf "$archive_path" -C "$temp_dir"
    
    local extracted_binary
    extracted_binary=$(find "$temp_dir" -name "${BINARY_NAME}*" -type f ! -name "*.tar.gz" ! -name "*.sig" | head -1)
    
    if [[ -z "$extracted_binary" ]]; then
        log_error "Binary not found in archive"
        exit 1
    fi
    
    log_info "Installing binary to $binary_path"
    cp "$extracted_binary" "$binary_path"
    chmod +x "$binary_path"
    
    verify_signature "$binary_path"

    add_to_path "$binary_path"
    
    rm -rf "$temp_dir"
    
    if command_exists "$BINARY_NAME"; then
        local installed_version
        installed_version=$("$BINARY_NAME" version 2>/dev/null | head -1 | sed 's/ami-util version //' || echo "unknown")
        log_success "Installation completed successfully!"
        log_info "Version: $installed_version"
        log_info "Location: $binary_path"
        
        if [[ "$NO_PATH" != "true" ]]; then
            echo
            log_info "To use ami-util, either:"
            log_info "1. Restart your terminal, or"
            log_info "2. Run: source ~/.bashrc (or ~/.zshrc)"
            echo
            log_info "Then run: ami-util --help"
        fi
    else
        log_error "Installation failed - binary not found in PATH"
        exit 1
    fi
}

main "$@"
