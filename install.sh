#!/usr/bin/env bash
set -euo pipefail

repo_slug="${REPO_SLUG:-ksinistr/openmeteo-cli}"
install_dir="${INSTALL_DIR:-$HOME/.local/bin}"
version="${VERSION:-latest}"

os="$(uname -s)"
arch="$(uname -m)"

case "${os}" in
  Linux) platform_os="linux" ;;
  Darwin) platform_os="darwin" ;;
  *)
    printf 'unsupported operating system: %s\n' "${os}" >&2
    exit 1
    ;;
esac

case "${arch}" in
  x86_64|amd64) platform_arch="amd64" ;;
  arm64|aarch64) platform_arch="arm64" ;;
  *)
    printf 'unsupported architecture: %s\n' "${arch}" >&2
    exit 1
    ;;
esac

asset_name="openmeteo-cli_${platform_os}_${platform_arch}"
checksums_name="checksums.txt"

if [ -n "${DOWNLOAD_BASE_URL:-}" ]; then
  asset_url="${DOWNLOAD_BASE_URL}/${asset_name}"
  checksums_url="${DOWNLOAD_BASE_URL}/${checksums_name}"
elif [ "${version}" = "latest" ]; then
  asset_url="https://github.com/${repo_slug}/releases/latest/download/${asset_name}"
  checksums_url="https://github.com/${repo_slug}/releases/latest/download/${checksums_name}"
else
  asset_url="https://github.com/${repo_slug}/releases/download/${version}/${asset_name}"
  checksums_url="https://github.com/${repo_slug}/releases/download/${version}/${checksums_name}"
fi

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "${tmp_dir}"
}
trap cleanup EXIT

binary_path="${tmp_dir}/openmeteo-cli"
checksums_path="${tmp_dir}/${checksums_name}"

curl -fsSL "${asset_url}" -o "${binary_path}"
curl -fsSL "${checksums_url}" -o "${checksums_path}"

expected_checksum="$(
  awk -v asset="${asset_name}" '
    {
      name = $2
      sub(/^.*\//, "", name)
      if (name == asset) {
        print $1
        exit
      }
    }
  ' "${checksums_path}"
)"
if [ -z "${expected_checksum}" ]; then
  printf 'missing checksum for %s\n' "${asset_name}" >&2
  exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
  actual_checksum="$(sha256sum "${binary_path}" | awk '{print $1}')"
elif command -v shasum >/dev/null 2>&1; then
  actual_checksum="$(shasum -a 256 "${binary_path}" | awk '{print $1}')"
else
  printf 'sha256sum or shasum is required to verify the downloaded binary\n' >&2
  exit 1
fi

if [ "${actual_checksum}" != "${expected_checksum}" ]; then
  printf 'checksum mismatch for %s\n' "${asset_name}" >&2
  exit 1
fi

mkdir -p "${install_dir}"
chmod +x "${binary_path}"
mv "${binary_path}" "${install_dir}/openmeteo-cli"

printf 'installed openmeteo-cli to %s/openmeteo-cli\n' "${install_dir}"
if ! printf '%s' ":${PATH}:" | grep -q ":${install_dir}:"; then
  printf 'add %s to PATH to run openmeteo-cli directly\n' "${install_dir}"
fi
