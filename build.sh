#!/usr/bin/env sh
set -eu

if [ "$#" -ne 1 ]; then
	echo "usage: $0 <version>" >&2
	exit 1
fi

version=$1

mkdir -p work
go build -ldflags "-X main.Version=${version}" -o work/dddns .
