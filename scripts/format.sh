#!/usr/bin/env bash
set -euo pipefail

gofumpt -w .
gci write --skip-generated -s standard -s default -s blank -s dot .
