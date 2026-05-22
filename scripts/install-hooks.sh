#!/usr/bin/env sh
# scripts/install-hooks.sh
# Configures Git to use the project's .githooks directory.
# Run once after cloning: sh scripts/install-hooks.sh

set -e

HOOKS_DIR=".githooks"

if [ ! -d "$HOOKS_DIR" ]; then
  echo "ERROR: $HOOKS_DIR directory not found."
  echo "Make sure you run this script from the repository root."
  exit 1
fi

git config core.hooksPath "$HOOKS_DIR"

# Ensure the hook scripts are executable
chmod +x "$HOOKS_DIR"/*

echo "Git hooks installed successfully."
echo "Hook path set to: $(git config core.hooksPath)"
