#!/bin/sh
# Sync windowing-backend C sources into vendor/ that 'go work vendor' misses.
#
# go work vendor only follows #include chains for the default build config,
# so raylib's RGFW sources (used by '-tags rgfw') are not copied into vendor/.
# Run this after every 'go work vendor'.
#
# Usage: tools/sync-vendor-extras.sh
set -e

MOD=$(GOWORK=off go list -m -f '{{.Dir}}' github.com/gen2brain/raylib-go/raylib)
RL_VENDOR=vendor/github.com/gen2brain/raylib-go/raylib/external

rm -rf "$RL_VENDOR/RGFW"
# Module cache files are read-only; normalize modes so future
# 'go work vendor' resets can delete them.
cp -r "$MOD/external/RGFW" "$RL_VENDOR/RGFW"
chmod -R u+w "$RL_VENDOR/RGFW"

echo "synced $RL_VENDOR/RGFW from $MOD"
