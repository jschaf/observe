#!/usr/bin/env bash
set -euo pipefail

echo '::group::Download Go'
goroot="./go"
mkdir -p "$goroot"
echo "$goroot" >> "$GITHUB_PATH"

# Read Go version from go.mod.
version=$(grep '^go ' "${GITHUB_WORKSPACE}/go.mod" | awk '{print $2}')
echo "Using Go version: $version"

# Detect platform and arch.
platform=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$(uname -m)" in
    x86_64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *) arch="amd64" ;;  # default fallback
esac

# Download.
archive_url="https://go.dev/dl/go${version}.${platform}-${arch}.tar.gz"
echo "Downloading from: $archive_url"
curl --fail --silent --show-error --location --write-out "%{stderr}Downloaded in %{time_total} seconds\n" "$archive_url" \
  | tar -xz --strip-components=1 -C "$goroot"

export PATH="$PATH:$goroot"
go version
echo '::endgroup::'

