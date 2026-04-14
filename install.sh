#!/usr/bin/env bash
set -euo pipefail

REPO="${REPO:-eremenko789/git-log-report-generator}"
BINARY_NAME="${BINARY_NAME:-git-html-report}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

if [[ "${VERSION}" == "latest" ]]; then
  RELEASE_API_URL="https://api.github.com/repos/${REPO}/releases/latest"
else
  RELEASE_API_URL="https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"
fi

detect_os() {
  case "$(uname -s)" in
    Linux) echo "linux" ;;
    Darwin) echo "darwin" ;;
    *)
      echo "Unsupported OS: $(uname -s)" >&2
      exit 1
      ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)
      echo "Unsupported architecture: $(uname -m)" >&2
      exit 1
      ;;
  esac
}

need_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Required command not found: $1" >&2
    exit 1
  fi
}

need_cmd curl
need_cmd tar

OS="$(detect_os)"
ARCH="$(detect_arch)"
ASSET_BASENAME="git-log-report-generator_${OS}_${ARCH}.tar.gz"

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

RELEASE_JSON="${TMP_DIR}/release.json"
ASSET_PATH="${TMP_DIR}/${ASSET_BASENAME}"
DOWNLOAD_URL=""

echo "Resolving release metadata from ${REPO}..."
curl -fsSL "${RELEASE_API_URL}" -o "${RELEASE_JSON}"

DOWNLOAD_URL="$(sed -n 's/.*"browser_download_url":[[:space:]]*"\([^"]*\)".*/\1/p' "${RELEASE_JSON}" | grep -F "${ASSET_BASENAME}" | head -n 1 || true)"

if [[ -z "${DOWNLOAD_URL}" ]]; then
  echo "Could not find asset ${ASSET_BASENAME} in release ${VERSION}" >&2
  exit 1
fi

echo "Downloading ${ASSET_BASENAME}..."
curl -fsSL "${DOWNLOAD_URL}" -o "${ASSET_PATH}"

echo "Extracting archive..."
tar -xzf "${ASSET_PATH}" -C "${TMP_DIR}"

if [[ ! -f "${TMP_DIR}/${BINARY_NAME}" ]]; then
  echo "Binary ${BINARY_NAME} not found in downloaded archive" >&2
  exit 1
fi

mkdir -p "${INSTALL_DIR}" 2>/dev/null || true

if [[ -w "${INSTALL_DIR}" ]]; then
  install -m 0755 "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
else
  if command -v sudo >/dev/null 2>&1; then
    sudo install -m 0755 "${TMP_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}"
  else
    echo "No write access to ${INSTALL_DIR} and sudo is unavailable." >&2
    echo "Try: INSTALL_DIR=\"${HOME}/.local/bin\" bash install.sh" >&2
    exit 1
  fi
fi

echo "Installed ${BINARY_NAME} to ${INSTALL_DIR}/${BINARY_NAME}"
echo "Run: ${BINARY_NAME} --version"
