#!/usr/bin/env bash
set -euo pipefail

echo '::group::Download Go'

# Read Go version from go.mod.
version=$(grep '^go ' "${GITHUB_WORKSPACE}/go.mod" | awk '{print $2}')

# Detect platform and arch.
platform=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in
    x86_64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *) arch="amd64" ;;  # default fallback
esac
echo "Go version: $version, platform: $platform, arch: $arch"

goroot="$RUNNER_TOOL_CACHE/go/$version/$arch"
mkdir -p "$goroot"
echo "$goroot/bin" >> "$GITHUB_PATH"

# Download.
archive_url="https://github.com/actions/go-versions/releases/download/1.24.5-16210585985/go-1.24.5-linux-x64.tar.gz"
#archive_url="https://go.dev/dl/go${version}.${platform}-${arch}.tar.gz"
echo "Downloading from: $archive_url"
curl --fail --silent --show-error --location --write-out "%{stderr}Downloaded in %{time_total} seconds\n" "$archive_url" \
  | tar -xz --strip-components=1 -C "$goroot"

"$goroot/bin/go" version
echo '::endgroup::'

