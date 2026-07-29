#!/usr/bin/env bash
set -euo pipefail

readonly atlas_version='v1.2.0'
readonly atlas_commit='47daa88aea519f7f4c4aab5adfde2beab9b10b13'
readonly atlas_repository='https://github.com/ariga/atlas.git'
readonly atlas_output="${PWD}/.tools/bin/atlas"

atlas_tmpdir="$(mktemp -d)"
trap 'rm -rf "$atlas_tmpdir"' EXIT

git clone --depth 1 --branch "$atlas_version" \
  "$atlas_repository" "$atlas_tmpdir/atlas"

if [ "$(git -C "$atlas_tmpdir/atlas" rev-parse HEAD)" != "$atlas_commit" ]; then
  echo "Atlas source commit does not match ${atlas_commit}" >&2
  exit 1
fi

mkdir -p "$(dirname "$atlas_output")"

(
  cd "$atlas_tmpdir/atlas/cmd/atlas"
  go build \
    -trimpath \
    -ldflags "-X ariga.io/atlas/cmd/atlas/internal/cmdapi.version=${atlas_version}" \
    -o "$atlas_output" \
    .
)

"$atlas_output" version
