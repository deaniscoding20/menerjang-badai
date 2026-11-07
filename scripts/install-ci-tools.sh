#!/usr/bin/env bash
# Linux amd64 only. Versions and SHA256 are reviewed pins, not floating latest.
set -euo pipefail
export PATH="$HOME/.local/bin:$PATH"
mkdir -p "$HOME/.local/bin"
work=$(mktemp -d "${TMPDIR:-/tmp}/mb-ci-tools.XXXXXX")
trap 'rm -rf "$work"' EXIT
curl --fail --silent --show-error --location --proto '=https' --tlsv1.2 \
  https://github.com/rhysd/actionlint/releases/download/v1.7.12/actionlint_1.7.12_linux_amd64.tar.gz \
  -o "$work/actionlint.tar.gz"
printf '%s  %s\n' 8aca8db96f1b94770f1b0d72b6dddcb1ebb8123cb3712530b08cc387b349a3d8 "$work/actionlint.tar.gz" | sha256sum --check
curl --fail --silent --show-error --location --proto '=https' --tlsv1.2 \
  https://github.com/gitleaks/gitleaks/releases/download/v8.30.1/gitleaks_8.30.1_linux_x64.tar.gz \
  -o "$work/gitleaks.tar.gz"
printf '%s  %s\n' 551f6fc83ea457d62a0d98237cbad105af8d557003051f41f3e7ca7b3f2470eb "$work/gitleaks.tar.gz" | sha256sum --check
tar -xzf "$work/actionlint.tar.gz" -C "$work" actionlint
tar -xzf "$work/gitleaks.tar.gz" -C "$work" gitleaks
install -m 0755 "$work/actionlint" "$HOME/.local/bin/actionlint"
install -m 0755 "$work/gitleaks" "$HOME/.local/bin/gitleaks"
actionlint --version
gitleaks version
