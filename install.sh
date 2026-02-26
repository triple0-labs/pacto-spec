#!/usr/bin/env bash
set -euo pipefail

REPO="${PACTO_REPO:-triple0-labs/pacto-spec}"
VERSION="${PACTO_VERSION:-latest}"

detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    linux|darwin) echo "$os" ;;
    *) echo "unsupported OS: $os" >&2; exit 1 ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m)"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
  esac
}

resolve_version() {
  if [[ "$VERSION" != "latest" ]]; then
    echo "${VERSION#v}"
    return
  fi
  local latest
  latest="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep -m1 '"tag_name"' | sed -E 's/.*"tag_name":[[:space:]]*"v?([^"]+)".*/\1/')"
  if [[ -z "$latest" ]]; then
    echo "failed to resolve latest release version from GitHub API" >&2
    exit 1
  fi
  echo "$latest"
}

sha256_file() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print tolower($1)}'
    return
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print tolower($1)}'
    return
  fi
  echo "missing checksum tool (sha256sum or shasum)" >&2
  exit 1
}

main() {
  local os arch version tag artifact tmpdir install_dir
  os="$(detect_os)"
  arch="$(detect_arch)"
  version="$(resolve_version)"
  tag="v${version}"
  artifact="pacto_${version}_${os}_${arch}.tar.gz"

  if [[ -n "${INSTALL_DIR:-}" ]]; then
    install_dir="$INSTALL_DIR"
  elif [[ -w "/usr/local/bin" ]]; then
    install_dir="/usr/local/bin"
  else
    install_dir="$HOME/.local/bin"
  fi

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "${tmpdir:-}"' EXIT

  echo "Installing pacto ${tag} for ${os}/${arch}..."
  curl -fsSL -o "${tmpdir}/${artifact}" "https://github.com/${REPO}/releases/download/${tag}/${artifact}"
  curl -fsSL -o "${tmpdir}/checksums.txt" "https://github.com/${REPO}/releases/download/${tag}/checksums.txt"

  local expected actual
  expected="$(awk -v a="$artifact" '$2 == a {print tolower($1)}' "${tmpdir}/checksums.txt")"
  if [[ -z "$expected" ]]; then
    echo "checksum entry not found for ${artifact}" >&2
    exit 1
  fi
  actual="$(sha256_file "${tmpdir}/${artifact}")"
  if [[ "$expected" != "$actual" ]]; then
    echo "checksum mismatch for ${artifact}" >&2
    exit 1
  fi

  tar -xzf "${tmpdir}/${artifact}" -C "${tmpdir}"
  mkdir -p "$install_dir"
  install -m 0755 "${tmpdir}/pacto" "${install_dir}/pacto"

  echo "Installed to ${install_dir}/pacto"
  if ! command -v pacto >/dev/null 2>&1; then
    echo "Note: ${install_dir} is not on PATH for this shell."
    echo "Add it with: export PATH=\"${install_dir}:\$PATH\""
  fi
  "${install_dir}/pacto" version
}

main "$@"
