#!/bin/bash
# Use of this source code is governed by Apache License 2.0
# that can be found in the LICENSE file.

# Read first argument and pass it as tag otherwise set test
[ -z "$1" ] && tag="test" || tag=$1

pkgs=(rpm deb apk)
archs=(amd64)

for pkg in "${pkgs[@]}"
do
  for arch in "${archs[@]}"
  do
	  echo "Building package ${pkg} ${arch} ${tag}"
	  env SEMVER="${tag}" nfpm package -f ./build/package/nfpm.yaml -t ./build/package/release -p "${pkg}"
  done
done
