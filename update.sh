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

INSTALL_DIR="/usr/local/bin"

header "Updateing Lockin binary"

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
