#!/bin/bash

set -euo pipefail

strip_carriage_return() {
	tr -d '\r'
}

resolve_build_version() {
	local toolchain_version
	local go_version

	toolchain_version="$(awk '/^toolchain / { sub(/^go/, "", $2); print $2; exit }' go.mod | strip_carriage_return)"
	if [ -n "$toolchain_version" ]; then
		printf '%s\n' "$toolchain_version"
		return
	fi

	go_version="$(awk '/^go / { print $2; exit }' go.mod | strip_carriage_return)"
	if [ -z "$go_version" ]; then
		echo "Unable to determine Go version from go.mod" >&2
		exit 1
	fi

	printf '%s\n' "$go_version"
}

BUILD_VERSION="$(resolve_build_version)"
BUILD_IMAGE="gotify/build:${BUILD_VERSION}-linux-amd64"

docker pull "$BUILD_IMAGE"
docker run --rm -v "$PWD/.:/proj" -w /proj "$BUILD_IMAGE" \
	go build -a -installsuffix cgo -ldflags "-w -s" -buildmode=plugin -o plugin/authentik-plugin-amd64.so /proj