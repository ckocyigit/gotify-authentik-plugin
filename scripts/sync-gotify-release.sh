#!/bin/bash

set -euo pipefail

if ! command -v curl >/dev/null 2>&1; then
	echo "curl is required" >&2
	exit 1
fi

if ! command -v go >/dev/null 2>&1; then
	echo "go is required" >&2
	exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

strip_carriage_return() {
	tr -d '\r'
}

resolve_latest_gotify_tag() {
	curl -fsSL https://api.github.com/repos/gotify/server/releases/latest \
		| sed -nE 's/.*"tag_name": "([^"]+)".*/\1/p' \
		| head -n 1
}

remove_toolchain_directive() {
	local file_path
	local tmp_file

	file_path="$1"
	tmp_file="$(mktemp)"
	grep -v '^toolchain ' "$file_path" > "$tmp_file" || true
	mv "$tmp_file" "$file_path"
	rm -f "$tmp_file"
}

TARGET_TAG="${1:-}"
if [ -z "$TARGET_TAG" ]; then
	TARGET_TAG="$(resolve_latest_gotify_tag)"
fi

if [ -z "$TARGET_TAG" ]; then
	echo "Unable to determine the target Gotify release tag" >&2
	exit 1
fi

TMP_GO_MOD="$(mktemp)"
trap 'rm -f "$TMP_GO_MOD"' EXIT

curl -fsSL "https://raw.githubusercontent.com/gotify/server/${TARGET_TAG}/go.mod" -o "$TMP_GO_MOD"

SERVER_GO_VERSION="$(awk '/^go / { print $2; exit }' "$TMP_GO_MOD" | strip_carriage_return)"
SERVER_TOOLCHAIN="$(awk '/^toolchain / { print $2; exit }' "$TMP_GO_MOD" | strip_carriage_return)"

if [ -z "$SERVER_GO_VERSION" ]; then
	echo "Unable to determine the Go version from Gotify server go.mod" >&2
	exit 1
fi

pushd "$REPO_ROOT" >/dev/null

PLUGIN_API_VERSION="$(awk '/github.com\/gotify\/plugin-api / { print $2; exit }' go.mod | strip_carriage_return)"
if [ -z "$PLUGIN_API_VERSION" ]; then
	echo "Unable to determine github.com/gotify/plugin-api version from go.mod" >&2
	exit 1
fi

go run github.com/gotify/plugin-api/cmd/gomod-cap@${PLUGIN_API_VERSION} -from "$TMP_GO_MOD" -to go.mod
go mod edit -go="$SERVER_GO_VERSION"

if [ -n "$SERVER_TOOLCHAIN" ]; then
	go mod edit -toolchain="$SERVER_TOOLCHAIN"
else
	remove_toolchain_directive go.mod
fi

go mod tidy

echo "Synced plugin dependencies to ${TARGET_TAG}" >&2
echo "Server go directive: ${SERVER_GO_VERSION}" >&2
if [ -n "$SERVER_TOOLCHAIN" ]; then
	echo "Server toolchain: ${SERVER_TOOLCHAIN}" >&2
fi

popd >/dev/null