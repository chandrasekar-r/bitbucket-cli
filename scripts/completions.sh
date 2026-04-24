#!/bin/bash
# Generate shell completion scripts for packaging with GoReleaser.
# Output goes to ./completions/ which is included in release archives and deb/rpm packages.
set -euo pipefail

rm -rf completions
mkdir completions

for shell in bash zsh fish; do
  echo "Generating $shell completion..."
  go run ./cmd/bb completion "$shell" > "completions/bb.$shell"
done

echo "Completions written to ./completions/"
