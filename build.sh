#!/usr/bin/env sh
set -eu

go build -ldflags "-X main.Version=indev-$(git rev-parse --short HEAD)" -o work/dddns .
