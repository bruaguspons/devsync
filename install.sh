#!/bin/bash
# Installs (or updates) the devsync binary and its bundled mise.toml
# and skills snapshots from the latest GitHub Release, mirroring the
# `curl -fsSL https://claude.ai/install.sh | bash` install pattern.
#
# Usage: curl -fsSL <raw-url>/install.sh | bash
set -euo pipefail

DEVSYNC_OWNER="bruaguspons"
DEVSYNC_REPO="devsync"
DEVSYNC_BIN_DIR="${HOME}/.local/bin"
DEVSYNC_DATA_DIR="${HOME}/.local/share/devsync"
DEVSYNC_BIN_NAME="devsync"

# validate_asset_filename rejects any asset filename that does not match
# the exact goreleaser-produced pattern for a supported platform, before
# any extraction happens. This is the threat-matrix guard against
# extracting an unexpected/malformed asset (e.g. path traversal, wrong
# OS/arch, wrong extension).
#
# Pattern: devsync_<version>_linux_<amd64|arm64>.tar.gz
validate_asset_filename() {
  local filename="$1"
  [[ -n "${filename}" ]] || return 1
  [[ "${filename}" =~ ^devsync_[0-9]+\.[0-9]+\.[0-9]+_linux_(amd64|arm64)\.tar\.gz$ ]]
}

detect_arch() {
  local machine
  machine="$(uname -m)"
  case "${machine}" in
    x86_64 | amd64) echo "amd64" ;;
    aarch64 | arm64) echo "arm64" ;;
    *)
      echo "devsync: unsupported architecture '${machine}'" >&2
      return 1
      ;;
  esac
}

detect_os() {
  local kernel
  kernel="$(uname -s)"
  case "${kernel}" in
    Linux) echo "linux" ;;
    *)
      echo "devsync: unsupported OS '${kernel}' (only linux is supported)" >&2
      return 1
      ;;
  esac
}

install_main() {
  local os arch api_url latest_json tag asset_name asset_url tmp_dir archive_path

  os="$(detect_os)"
  arch="$(detect_arch)"

  api_url="https://api.github.com/repos/${DEVSYNC_OWNER}/${DEVSYNC_REPO}/releases/latest"
  latest_json="$(curl -fsSL "${api_url}")"

  tag="$(printf '%s' "${latest_json}" | grep -o '"tag_name"[[:space:]]*:[[:space:]]*"[^"]*"' | head -1 | sed -E 's/.*"([^"]+)"$/\1/')"
  if [[ -z "${tag}" ]]; then
    echo "devsync: could not determine latest release tag" >&2
    return 1
  fi

  local version="${tag#v}"
  asset_name="devsync_${version}_${os}_${arch}.tar.gz"

  if ! validate_asset_filename "${asset_name}"; then
    echo "devsync: refusing to install unexpected/malformed asset name '${asset_name}'" >&2
    return 1
  fi

  asset_url="https://github.com/${DEVSYNC_OWNER}/${DEVSYNC_REPO}/releases/download/${tag}/${asset_name}"

  tmp_dir="$(mktemp -d)"
  trap 'rm -rf "${tmp_dir}"' EXIT

  archive_path="${tmp_dir}/${asset_name}"
  curl -fsSL -o "${archive_path}" "${asset_url}"

  mkdir -p "${DEVSYNC_BIN_DIR}" "${DEVSYNC_DATA_DIR}"
  tar -xzf "${archive_path}" -C "${tmp_dir}"

  if [[ ! -f "${tmp_dir}/${DEVSYNC_BIN_NAME}" ]]; then
    echo "devsync: extracted archive is missing the expected binary" >&2
    return 1
  fi
  if [[ ! -f "${tmp_dir}/mise.toml.snapshot" ]]; then
    echo "devsync: extracted archive is missing the expected mise.toml.snapshot" >&2
    return 1
  fi
  if [[ ! -d "${tmp_dir}/skills.snapshot" ]]; then
    echo "devsync: extracted archive is missing the expected skills.snapshot directory" >&2
    return 1
  fi

  install -m 0755 "${tmp_dir}/${DEVSYNC_BIN_NAME}" "${DEVSYNC_BIN_DIR}/${DEVSYNC_BIN_NAME}"
  install -m 0644 "${tmp_dir}/mise.toml.snapshot" "${DEVSYNC_DATA_DIR}/mise.toml.snapshot"
  # rm-then-mv (instead of rm-then-cp) narrows the failure window: an
  # interrupted mv fails atomically as "old removed, new absent," never
  # "half-copied." Not a true atomic rename if tmp_dir and
  # DEVSYNC_DATA_DIR are on different filesystems (mv degrades to
  # copy+delete internally in that case) — accepted residual risk.
  rm -rf "${DEVSYNC_DATA_DIR}/skills.snapshot"
  mv "${tmp_dir}/skills.snapshot" "${DEVSYNC_DATA_DIR}/skills.snapshot"

  echo "devsync: installed ${tag} to ${DEVSYNC_BIN_DIR}/${DEVSYNC_BIN_NAME}"
}

# Allow this script to be `source`d for unit testing its functions
# (e.g. validate_asset_filename) without running the network install flow.
if [[ "${DEVSYNC_SOURCE_ONLY:-}" != "1" ]]; then
  install_main "$@"
fi
