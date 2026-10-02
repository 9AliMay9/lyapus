#!/usr/bin/env bash
set -euo pipefail

readonly atlas_version='v1.3.0'
readonly atlas_build_version='v1.3.0-lyapus.1'
readonly atlas_commit='9a6bc601212130aaaefcbc8dd36c710baf9716ff'
readonly atlas_repository='https://github.com/ariga/atlas.git'
readonly atlas_output="${LYAPUS_ATLAS_OUTPUT:-${PWD}/.tools/bin/atlas}"

atlas_script_dir="$(
  cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd
)"
readonly atlas_patch="${atlas_script_dir}/patches/atlas-v1.3.0-dependencies.patch"
readonly atlas_patch_sha256='e9b0ec5264bfd9e33fa0f31d691aacbb53c2ecb59b1385cb011bf19295142638'

if ! printf '%s  %s\n' "$atlas_patch_sha256" "$atlas_patch" |
  sha256sum --check --status; then
  echo "Atlas dependency patch verification failed" >&2
  exit 1
fi

case "$atlas_output" in
  /*) ;;
  *)
    echo "Atlas output path must be absolute" >&2
    exit 1
    ;;
esac

if [ -d "$atlas_output" ] || [ -L "$atlas_output" ]; then
  echo "Atlas output must not be a directory or symbolic link" >&2
  exit 1
fi

atlas_output_dir="$(dirname "$atlas_output")"
mkdir -p "$atlas_output_dir"

atlas_tmpdir="$(mktemp -d "${atlas_output_dir}/.atlas-build.XXXXXX")"
trap 'rm -rf -- "$atlas_tmpdir"' EXIT

git clone --depth 1 --branch "$atlas_version" \
  "$atlas_repository" "$atlas_tmpdir/source"

if [ "$(git -C "$atlas_tmpdir/source" rev-parse HEAD)" != "$atlas_commit" ]; then
  echo "Atlas source commit does not match ${atlas_commit}" >&2
  exit 1
fi

git -C "$atlas_tmpdir/source" apply --check "$atlas_patch"
git -C "$atlas_tmpdir/source" apply "$atlas_patch"

(
  cd "$atlas_tmpdir/source/cmd/atlas"
  GOTOOLCHAIN=local go build \
    -mod=readonly \
    -trimpath \
    -ldflags "-X ariga.io/atlas/cmd/atlas/internal/cmdapi.version=${atlas_build_version}" \
    -o "$atlas_tmpdir/atlas" \
    .
)

atlas_version_output="$(
  ATLAS_NO_UPDATE_NOTIFIER=1 "$atlas_tmpdir/atlas" version
)"
printf '%s\n' "$atlas_version_output"

if ! printf '%s\n' "$atlas_version_output" |
  grep -Fqx -- "atlas community version ${atlas_build_version}"; then
  echo "Atlas version verification failed" >&2
  exit 1
fi

mv -fT -- "$atlas_tmpdir/atlas" "$atlas_output"
printf 'Installed Atlas to %s\n' "$atlas_output"
