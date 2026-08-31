#!/usr/bin/env sh
set -eu

go build -ldflags "-X main.Version=indev-$(git describe --tags --always --dirty)" -o work/dddns .
