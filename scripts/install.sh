#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! command -v go >/dev/null 2>&1; then
	echo "error: go is required" >&2
	exit 1
fi

echo "Installing fonts..."
"${root}/scripts/install-fonts.sh"

echo "Installing translate via go install..."
go install github.com/fuba/translate/cmd/translate@latest

gobin="$(go env GOBIN 2>/dev/null || true)"
gopath="$(go env GOPATH 2>/dev/null || true)"
if [ -n "${gobin}" ]; then
	bin_dir="${gobin}"
else
	bin_dir="${gopath}/bin"
fi

if [ -n "${bin_dir}" ] && [ -d "${bin_dir}" ]; then
	if ! echo ":$PATH:" | grep -q ":${bin_dir}:"; then
		echo "warning: ${bin_dir} is not in PATH"
		echo "Add this to your shell profile:"
		echo "  export PATH=\"${bin_dir}:\$PATH\""
	fi
fi

echo "Done."
