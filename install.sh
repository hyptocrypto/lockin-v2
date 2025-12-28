#!/bin/bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
BOLD='\033[1m'
DIM='\033[2m'
NC='\033[0m' # No Color

# Styled output helpers
info() { echo -e "${CYAN}$1${NC}"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
header() { echo -e "${BOLD}${MAGENTA}$1${NC}"; }
dim() { echo -e "${DIM}$1${NC}"; }
step() { echo -e "${BLUE}→${NC} $1..."; }

CONFIG_DIR="$HOME/.lockin"
CONFIG_FILE="$CONFIG_DIR/config.yaml"
INSTALL_DIR="/usr/local/bin"

echo -e "${MAGENTA}
    __               __   ____    
   / /   ____  _____/ /__/  _/___ 
  / /   / __ \/ ___/ //_// // __ \\
 / /___/ /_/ / /__/ ,< _/ // / / /
/_____/\____/\___/_/|_/___/_/ /_/ 
${NC}"
header "LockIn Installer"
echo

# Create config directory
mkdir -p "$CONFIG_DIR"
chmod 700 "$CONFIG_DIR"

# SMB Configuration
header "SMB Sync Configuration"
dim "Sync your vault across devices on your local network."
dim "(Leave blank to disable SMB sync)"
echo

echo -en "${BOLD}Enable SMB sync? [y/N]:${NC} "
read enable_smb

if [[ "$enable_smb" =~ ^[Yy]$ ]]; then
    SMB_ENABLED="true"
    echo
    
    echo -en "${CYAN}SMB Host${NC} ${DIM}(e.g., 192.168.1.100)${NC}: "
    read smb_host
    echo -en "${CYAN}SMB Port${NC} ${DIM}[445]${NC}: "
    read smb_port
    smb_port="${smb_port:-445}"
    echo -en "${CYAN}SMB Share name${NC}: "
    read smb_share
    echo -en "${CYAN}SMB Username${NC}: "
    read smb_user
    echo -en "${CYAN}SMB Password${NC}: "
    read -s smb_password
    echo
else
    SMB_ENABLED="false"
    smb_host=""
    smb_port="445"
    smb_share=""
    smb_user=""
    smb_password=""
fi

# Write config file
echo
step "Writing config to $CONFIG_FILE"

cat > "$CONFIG_FILE" << EOF
enabled: $SMB_ENABLED
host: "$smb_host"
port: "$smb_port"
share: "$smb_share"
user: "$smb_user"
password: "$smb_password"
EOF

chmod 600 "$CONFIG_FILE"
success "Config saved"

# Build binary
echo
step "Building LockIn"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"
go build -o lockin .
success "Build complete"

# Install binary
echo
step "Installing to $INSTALL_DIR"

if [[ -w "$INSTALL_DIR" ]]; then
    mv lockin "$INSTALL_DIR/"
else
    warn "Need sudo to install to $INSTALL_DIR"
    sudo mv lockin "$INSTALL_DIR/"
fi

chmod 755 "$INSTALL_DIR/lockin"
success "Installed to $INSTALL_DIR/lockin"

echo
success "Installation complete!"
info "Run 'lockin' to get started."
