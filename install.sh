#!/bin/bash

# Octopus CLI Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/VibeAny/octopus-cli/main/install.sh | bash

set -e  # Exit on any error

# Check if we're running in bash, if not try to re-execute with bash
if [ -z "$BASH_VERSION" ]; then
    if command -v bash >/dev/null 2>&1; then
        exec bash "$0" "$@"
    fi
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
NC='\033[0m' # No Color

# Constants
GITHUB_REPO="VibeAny/octopus-cli"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="octopus"

# Banner
print_banner() {
    printf "${PURPLE}\n"
    echo "  ╔═══════════════════════════════════════╗"
    echo "  ║                                       ║"
    echo "  ║           🐙 Octopus CLI              ║"
    echo "  ║                                       ║"
    echo "  ║     Dynamic Claude Code API Tool      ║"
    echo "  ║                                       ║"
    echo "  ╚═══════════════════════════════════════╝"
    printf "${NC}\n"
}

# Logging functions
log_info() {
    printf "${BLUE}[INFO]${NC} %s\n" "$1"
}

log_success() {
    printf "${GREEN}[SUCCESS]${NC} %s\n" "$1"
}

log_warn() {
    printf "${YELLOW}[WARN]${NC} %s\n" "$1"
}

log_error() {
    printf "${RED}[ERROR]${NC} %s\n" "$1"
}

# Detect OS and architecture
detect_platform() {
    local os arch

    # Detect OS
    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="macos" ;;
        CYGWIN*|MINGW*|MSYS*) os="windows" ;;
        *)          
            log_error "Unsupported operating system: $(uname -s)"
            exit 1 ;;
    esac

    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        arm64|aarch64)  arch="arm64" ;;
        i386|i686)      arch="386" ;;
        *)              
            log_error "Unsupported architecture: $(uname -m)"
            exit 1 ;;
    esac

    echo "${os}-${arch}"
}

# Get the latest release version
get_latest_version() {
    local version
    
    if command -v curl >/dev/null 2>&1; then
        version=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        version=$(wget -qO- "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi

    if [ -z "$version" ]; then
        log_error "Failed to get latest version"
        exit 1
    fi

    echo "$version"
}

# Get the download URL for the binary
get_download_url() {
    local version="$1"
    local platform="$2"
    local extension=""
    
    if [[ "$platform" == "windows-"* ]]; then
        extension=".exe"
    fi

    # Get the actual binary filename from GitHub releases
    local release_data
    if command -v curl >/dev/null 2>&1; then
        release_data=$(curl -fsSL "https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${version}")
    else
        release_data=$(wget -qO- "https://api.github.com/repos/${GITHUB_REPO}/releases/tags/${version}")
    fi

    # Find the binary that matches our platform pattern
    local binary_name
    binary_name=$(echo "$release_data" | grep '"name":' | grep "$platform" | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$binary_name" ]; then
        log_error "No binary found for platform: $platform"
        log_info "Available binaries:"
        echo "$release_data" | grep '"name":' | grep -E "octopus.*\.(exe)?$" | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^/  - /'
        exit 1
    fi

    echo "https://github.com/${GITHUB_REPO}/releases/download/${version}/${binary_name}"
}

# Download and install the binary
install_binary() {
    local download_url="$1"
    local temp_file="/tmp/octopus-installer-$$"
    
    log_info "Downloading from: $download_url"
    
    # Download the binary
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$download_url" -o "$temp_file"
    else
        wget -q "$download_url" -O "$temp_file"
    fi
    
    # Make it executable
    chmod +x "$temp_file"
    
    # Move to install directory
    log_info "Installing to $INSTALL_DIR/$BINARY_NAME"
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$temp_file" "$INSTALL_DIR/$BINARY_NAME"
    else
        log_warn "Need sudo privileges to install to $INSTALL_DIR"
        sudo mv "$temp_file" "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    log_success "Octopus CLI installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local version
        version=$("$BINARY_NAME" version 2>/dev/null || echo "unknown")
        log_success "Installation verified! Version: $version"
        return 0
    else
        log_error "Installation verification failed. $BINARY_NAME not found in PATH."
        log_info "You may need to restart your terminal or add $INSTALL_DIR to your PATH."
        return 1
    fi
}

# Show post-install instructions
show_instructions() {
    echo ""
    printf "${WHITE}🎉 Installation Complete!${NC}\n"
    echo ""
    printf "${CYAN}Quick Start:${NC}\n"
    echo "  1. Add API configuration:"
    echo "     octopus config add official https://api.anthropic.com sk-ant-your-key"
    echo ""
    echo "  2. Start the proxy service:"
    echo "     octopus start"
    echo ""
    echo "  3. Configure Claude Code to use http://localhost:8080"
    echo ""
    echo "  4. Switch APIs dynamically:"
    echo "     octopus config switch <api-name>"
    echo ""
    printf "${CYAN}More Information:${NC}\n"
    echo "  • Documentation: https://github.com/${GITHUB_REPO}/blob/main/README.md"
    echo "  • 中文文档: https://github.com/${GITHUB_REPO}/blob/main/README_CN.md"
    echo "  • Get help: octopus --help"
    echo ""
    printf "${GREEN}Happy coding with Octopus CLI! 🐙${NC}\n"
}

# Cleanup function
cleanup() {
    local temp_files="/tmp/octopus-installer-*"
    rm -f $temp_files 2>/dev/null || true
}

# Main installation function
main() {
    # Set up cleanup trap
    trap cleanup EXIT

    print_banner
    
    log_info "Starting Octopus CLI installation..."
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
        log_error "Neither curl nor wget is available. Please install one of them first."
        exit 1
    fi

    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # Get latest version
    local version
    log_info "Fetching latest release version..."
    version=$(get_latest_version)
    log_info "Latest version: $version"
    
    # Get download URL
    local download_url
    download_url=$(get_download_url "$version" "$platform")
    
    # Install binary
    install_binary "$download_url"
    
    # Verify installation
    if verify_installation; then
        show_instructions
    else
        exit 1
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Octopus CLI Installation Script"
        echo ""
        echo "Usage: curl -fsSL https://raw.githubusercontent.com/VibeAny/octopus-cli/main/install.sh | bash"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --version      Show version information"
        echo ""
        echo "Environment Variables:"
        echo "  INSTALL_DIR    Installation directory (default: /usr/local/bin)"
        echo ""
        exit 0
        ;;
    --version)
        echo "Octopus CLI Installer v1.0.0"
        exit 0
        ;;
esac

# Run main installation
main "$@"