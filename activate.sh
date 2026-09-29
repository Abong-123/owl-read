#!/bin/bash
export GOMODCACHE="$PWD/.venv-go"
export GOFLAGS="-mod=vendor"
echo "Go environment activated"
echo "   GOMODCACHE=$GOMODCACHE"
