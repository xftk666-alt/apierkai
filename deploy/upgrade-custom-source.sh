#!/usr/bin/env bash

set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

REPO_URL="${REPO_URL:-https://github.com/xftk666-alt/apierkai.git}"
BRANCH="${BRANCH:-main}"
INSTALL_DIR="${INSTALL_DIR:-/opt/sub2api}"
DEPLOY_DIR="${INSTALL_DIR}/deploy"

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

require_root() {
    if [ "$(id -u)" -ne 0 ]; then
        print_error "Please run this script as root or with sudo."
        exit 1
    fi
}

main() {
    require_root

    if ! command_exists git; then
        print_error "git is required."
        exit 1
    fi

    if ! command_exists docker; then
        print_error "docker is required."
        exit 1
    fi

    if ! docker compose version >/dev/null 2>&1; then
        print_error "docker compose is required."
        exit 1
    fi

    if [ ! -d "${INSTALL_DIR}/.git" ]; then
        print_error "Repository not found at ${INSTALL_DIR}"
        exit 1
    fi

    print_info "Updating repository..."
    git -C "${INSTALL_DIR}" remote set-url origin "${REPO_URL}"
    git -C "${INSTALL_DIR}" fetch origin "${BRANCH}"
    git -C "${INSTALL_DIR}" checkout "${BRANCH}"
    git -C "${INSTALL_DIR}" pull --ff-only origin "${BRANCH}"

    print_info "Rebuilding and restarting services..."
    cd "${DEPLOY_DIR}"
    docker compose -f docker-compose.local.yml -f docker-compose.source.yml up -d --build

    print_success "Upgrade completed."
    echo "Current commit: $(git -C "${INSTALL_DIR}" rev-parse --short HEAD)"
}

main "$@"
