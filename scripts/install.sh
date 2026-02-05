#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v go >/dev/null 2>&1; then
	echo "error: go is required" >&2
	exit 1
fi

echo "Installing fonts..."
"${root}/scripts/install-fonts.sh"

gobin="$(go env GOBIN 2>/dev/null || true)"
gopath="$(go env GOPATH 2>/dev/null || true)"
if [ -n "${gobin}" ]; then
	bin_dir="${gobin}"
else
	bin_dir="${gopath}/bin"
fi

if [ -z "${bin_dir}" ]; then
	echo "error: failed to resolve install directory" >&2
	exit 1
fi

mkdir -p "${bin_dir}"
echo "Building translate..."
go build -o "${bin_dir}/translate" ./cmd/translate

config_dir="${XDG_CONFIG_HOME:-$HOME/.config}/translate"
config_path="${config_dir}/config.json"
if [ ! -f "${config_path}" ]; then
	mkdir -p "${config_dir}"
	cat > "${config_path}" <<'JSON'
{
  "base_url": "",
  "model": "gpt-oss-20b",
  "from": "",
  "to": "",
  "format": "",
  "timeout_seconds": 120,
  "max_chars": 2000,
  "endpoint": "completion",
  "passphrase_ttl_seconds": 600,
  "pdf_font": ""
}
JSON
	echo "Wrote config template to ${config_path}"
	echo "Set base_url in the config to your API endpoint."
else
	echo "Config exists at ${config_path}"
fi

if [ -d "${bin_dir}" ] && ! echo ":$PATH:" | grep -q ":${bin_dir}:"; then
	echo "warning: ${bin_dir} is not in PATH"
	echo "Add this to your shell profile:"
	echo "  export PATH=\"${bin_dir}:\$PATH\""
fi

echo "Done."
