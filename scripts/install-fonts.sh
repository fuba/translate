#!/usr/bin/env bash
set -euo pipefail

config_dir="${XDG_CONFIG_HOME:-$HOME/.config}/translate"
font_dir="${config_dir}/fonts"
mkdir -p "${font_dir}"

expected_fonts=(
	"LINESeedJP-Regular.ttf"
	"LINESeedJP-Bold.ttf"
	"LINESeedJP-ExtraBold.ttf"
	"LINESeedJP-Thin.ttf"
)

missing=0
for font in "${expected_fonts[@]}"; do
	if [ ! -f "${font_dir}/${font}" ]; then
		missing=1
		break
	fi
done

if [ "${missing}" -eq 0 ]; then
	echo "LINE Seed JP fonts already installed in ${font_dir}"
	exit 0
fi

download_url=""
if command -v gh >/dev/null 2>&1; then
	download_url="$(gh api /repos/line/seed/releases/latest --jq '.assets[] | select(.name|test("seed-.*\\.zip|LINESeed-fonts\\.zip")) | .browser_download_url' | head -n1 || true)"
fi

if [ -z "${download_url}" ]; then
	download_url="https://github.com/line/seed/releases/download/v20251119/seed-v20251119.zip"
fi

tmp_dir="$(mktemp -d)"
cleanup() {
	rm -rf "${tmp_dir}"
}
trap cleanup EXIT

archive_path="${tmp_dir}/seed.zip"
curl -L -o "${archive_path}" "${download_url}"

unzip -q "${archive_path}" -d "${tmp_dir}/extracted"

ttf_dir="$(find "${tmp_dir}/extracted" -type d -path "*/LINESeedJP/fonts/ttf" -print -quit)"
if [ -z "${ttf_dir}" ]; then
	echo "error: LINESeedJP ttf directory not found in archive" >&2
	exit 1
fi

cp "${ttf_dir}"/*.ttf "${font_dir}/"
echo "Installed LINE Seed JP fonts to ${font_dir}"
