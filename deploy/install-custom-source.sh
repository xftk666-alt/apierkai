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

BIND_HOST_VALUE="${BIND_HOST:-0.0.0.0}"
SERVER_PORT_VALUE="${SERVER_PORT:-8080}"
TZ_VALUE="${TZ:-Asia/Shanghai}"
NPM_REGISTRY_VALUE="${NPM_REGISTRY:-https://registry.npmmirror.com}"
PNPM_VERSION_VALUE="${PNPM_VERSION:-9}"

POSTGRES_USER_VALUE="${POSTGRES_USER:-xwqwert}"
POSTGRES_DB_VALUE="${POSTGRES_DB:-xwqwert}"
POSTGRES_PASSWORD_VALUE="${POSTGRES_PASSWORD:-xw123456}"
REDIS_PASSWORD_VALUE="${REDIS_PASSWORD:-}"

ADMIN_EMAIL_VALUE="${ADMIN_EMAIL:-admin@sub2api.local}"
ADMIN_PASSWORD_VALUE="${ADMIN_PASSWORD:-}"
JWT_SECRET_VALUE="${JWT_SECRET:-}"
TOTP_ENCRYPTION_KEY_VALUE="${TOTP_ENCRYPTION_KEY:-}"

CURRENT_TIME="$(date +%Y%m%d%H%M%S)"
SUB2API_IMAGE_TAG_VALUE="${SUB2API_IMAGE_TAG:-custom-${CURRENT_TIME}}"
SUB2API_BUILD_DATE_VALUE="${SUB2API_BUILD_DATE:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
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

detect_package_manager() {
    if command_exists apt-get; then
        PACKAGE_MANAGER="apt-get"
        return
    fi

    if command_exists dnf; then
        PACKAGE_MANAGER="dnf"
        return
    fi

    if command_exists yum; then
        PACKAGE_MANAGER="yum"
        return
    fi

    print_error "Unsupported system package manager. Supported: apt-get / dnf / yum"
    exit 1
}

install_base_dependencies() {
    detect_package_manager

    print_info "Installing required system packages..."

    case "${PACKAGE_MANAGER}" in
        apt-get)
            apt-get update
            apt-get install -y ca-certificates curl git grep sed coreutils openssl gnupg
            ;;
        dnf)
            dnf install -y ca-certificates curl git grep sed coreutils openssl
            ;;
        yum)
            yum install -y ca-certificates curl git grep sed coreutils openssl
            ;;
    esac
}

ensure_docker_service() {
    if command_exists systemctl; then
        systemctl enable docker >/dev/null 2>&1 || true
        systemctl start docker
    fi
}

install_compose_for_apt() {
    if docker compose version >/dev/null 2>&1; then
        return
    fi

    if apt-cache show docker-compose-plugin >/dev/null 2>&1; then
        apt-get install -y docker-compose-plugin
        return
    fi

    if command_exists docker-compose; then
        return
    fi

    if apt-cache show docker-compose >/dev/null 2>&1; then
        apt-get install -y docker-compose
        return
    fi

    print_error "Neither docker compose plugin nor docker-compose is available."
    exit 1
}

setup_docker_apt_repo() {
    local codename
    local architecture

    codename=""
    architecture="$(dpkg --print-architecture)"

    if [ -f /etc/os-release ]; then
        # shellcheck disable=SC1091
        . /etc/os-release
        codename="${VERSION_CODENAME:-}"
    fi

    if [ -z "${codename}" ]; then
        codename="focal"
    fi

    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
    chmod a+r /etc/apt/keyrings/docker.asc
    echo "deb [arch=${architecture} signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu ${codename} stable" >/etc/apt/sources.list.d/docker.list
}

install_docker_for_apt() {
    export DEBIAN_FRONTEND=noninteractive

    if ! docker compose version >/dev/null 2>&1; then
        print_info "Installing Docker Engine and Compose plugin from Docker apt repository..."
        apt-get update
        apt-get install -y ca-certificates curl gnupg
        setup_docker_apt_repo
        apt-get update
        apt-get remove -y docker.io docker-doc docker-compose docker-compose-v2 podman-docker containerd runc >/dev/null 2>&1 || true
        apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
        print_success "Docker Engine and Compose plugin installed."
    elif command_exists docker; then
        print_info "Docker and Compose plugin already installed."
    fi

    ensure_docker_service

    if ! docker compose version >/dev/null 2>&1; then
        print_warning "Docker Compose plugin is unavailable, trying legacy docker-compose..."
        install_compose_for_apt
    fi
}

resolve_compose_command() {
    if docker compose version >/dev/null 2>&1; then
        COMPOSE_CMD=("docker" "compose")
        return
    fi

    if command_exists docker-compose; then
        COMPOSE_CMD=("docker-compose")
        return
    fi

    print_error "Neither 'docker compose' nor 'docker-compose' is available."
    exit 1
}

install_docker_if_needed() {
    case "${PACKAGE_MANAGER}" in
        apt-get)
            install_docker_for_apt
            ;;
        dnf)
            if ! command_exists docker; then
                print_info "Docker not found, installing from dnf..."
                dnf install -y docker docker-compose-plugin || dnf install -y moby-engine docker-compose
            else
                print_info "Docker already installed."
            fi
            ensure_docker_service
            ;;
        yum)
            if ! command_exists docker; then
                print_info "Docker not found, installing from yum..."
                yum install -y docker docker-compose-plugin || yum install -y docker docker-compose
            else
                print_info "Docker already installed."
            fi
            ensure_docker_service
            ;;
    esac

    resolve_compose_command
}

generate_hex_secret() {
    openssl rand -hex 32
}

generate_password() {
    openssl rand -hex 10
}

clone_or_update_repo() {
    if [ -d "${INSTALL_DIR}/.git" ]; then
        print_info "Repository already exists, updating source..."
        git -C "${INSTALL_DIR}" remote set-url origin "${REPO_URL}"
        git -C "${INSTALL_DIR}" fetch origin "${BRANCH}"
        git -C "${INSTALL_DIR}" checkout "${BRANCH}"
        git -C "${INSTALL_DIR}" pull --ff-only origin "${BRANCH}"
    else
        print_info "Cloning source repository..."
        mkdir -p "$(dirname "${INSTALL_DIR}")"
        git clone --branch "${BRANCH}" "${REPO_URL}" "${INSTALL_DIR}"
    fi
}

upsert_env() {
    local key="$1"
    local value="$2"
    local file="$3"
    local temp_file

    temp_file="$(mktemp)"

    awk -v key="${key}" -v value="${value}" '
        BEGIN {
            updated = 0
        }
        index($0, key "=") == 1 {
            print key "=" value
            updated = 1
            next
        }
        {
            print
        }
        END {
            if (!updated) {
                print key "=" value
            }
        }
    ' "${file}" >"${temp_file}"

    mv "${temp_file}" "${file}"
}

prepare_env_file() {
    local env_file="${DEPLOY_DIR}/.env"
    local env_example_file="${DEPLOY_DIR}/.env.example"
    local commit_short

    if [ ! -f "${env_example_file}" ]; then
        print_error "Missing ${env_example_file}"
        exit 1
    fi

    if [ ! -f "${env_file}" ]; then
        cp "${env_example_file}" "${env_file}"
    fi

    if [ -z "${JWT_SECRET_VALUE}" ]; then
        JWT_SECRET_VALUE="$(generate_hex_secret)"
    fi

    if [ -z "${TOTP_ENCRYPTION_KEY_VALUE}" ]; then
        TOTP_ENCRYPTION_KEY_VALUE="$(generate_hex_secret)"
    fi

    if [ -z "${ADMIN_PASSWORD_VALUE}" ]; then
        ADMIN_PASSWORD_VALUE="$(generate_password)"
    fi

    commit_short="$(git -C "${INSTALL_DIR}" rev-parse --short HEAD)"

    upsert_env "BIND_HOST" "${BIND_HOST_VALUE}" "${env_file}"
    upsert_env "SERVER_PORT" "${SERVER_PORT_VALUE}" "${env_file}"
    upsert_env "TZ" "${TZ_VALUE}" "${env_file}"
    upsert_env "NPM_REGISTRY" "${NPM_REGISTRY_VALUE}" "${env_file}"
    upsert_env "PNPM_VERSION" "${PNPM_VERSION_VALUE}" "${env_file}"
    upsert_env "POSTGRES_USER" "${POSTGRES_USER_VALUE}" "${env_file}"
    upsert_env "POSTGRES_PASSWORD" "${POSTGRES_PASSWORD_VALUE}" "${env_file}"
    upsert_env "POSTGRES_DB" "${POSTGRES_DB_VALUE}" "${env_file}"
    upsert_env "REDIS_PASSWORD" "${REDIS_PASSWORD_VALUE}" "${env_file}"
    upsert_env "ADMIN_EMAIL" "${ADMIN_EMAIL_VALUE}" "${env_file}"
    upsert_env "ADMIN_PASSWORD" "${ADMIN_PASSWORD_VALUE}" "${env_file}"
    upsert_env "JWT_SECRET" "${JWT_SECRET_VALUE}" "${env_file}"
    upsert_env "TOTP_ENCRYPTION_KEY" "${TOTP_ENCRYPTION_KEY_VALUE}" "${env_file}"
    upsert_env "SUB2API_IMAGE_TAG" "${SUB2API_IMAGE_TAG_VALUE}" "${env_file}"
    upsert_env "SUB2API_BUILD_VERSION" "${BRANCH}" "${env_file}"
    upsert_env "SUB2API_BUILD_COMMIT" "${commit_short}" "${env_file}"
    upsert_env "SUB2API_BUILD_DATE" "${SUB2API_BUILD_DATE_VALUE}" "${env_file}"

    chmod 600 "${env_file}"
}

prepare_directories() {
    mkdir -p "${DEPLOY_DIR}/data" "${DEPLOY_DIR}/postgres_data" "${DEPLOY_DIR}/redis_data"
}

start_services() {
    print_info "Starting apierkai services..."
    cd "${DEPLOY_DIR}"
    "${COMPOSE_CMD[@]}" -f docker-compose.local.yml -f docker-compose.source.yml up -d --build
}

show_summary() {
    echo ""
    echo "=================================================="
    echo "  apierkai one-click installation completed"
    echo "=================================================="
    echo ""
    echo "Install directory: ${INSTALL_DIR}"
    echo "Repository:        ${REPO_URL}"
    echo "Branch:            ${BRANCH}"
    echo "Panel URL:         http://<server-ip>:${SERVER_PORT_VALUE}"
    echo ""
    echo "Database user:     ${POSTGRES_USER_VALUE}"
    echo "Database name:     ${POSTGRES_DB_VALUE}"
    echo "Database password: ${POSTGRES_PASSWORD_VALUE}"
    echo "Admin email:       ${ADMIN_EMAIL_VALUE}"
    echo "Admin password:    ${ADMIN_PASSWORD_VALUE}"
    echo ""
    echo "Useful commands:"
    echo "  cd ${DEPLOY_DIR}"
    echo "  ${COMPOSE_CMD[*]} -f docker-compose.local.yml -f docker-compose.source.yml ps"
    echo "  ${COMPOSE_CMD[*]} -f docker-compose.local.yml -f docker-compose.source.yml logs -f sub2api"
    echo ""
    print_warning "Please save the admin password and .env file securely."
}

main() {
    require_root
    install_base_dependencies
    install_docker_if_needed
    clone_or_update_repo
    prepare_env_file
    prepare_directories
    start_services
    show_summary
}

main "$@"
