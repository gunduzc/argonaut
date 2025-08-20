#!/bin/sh
set -e

VERSION="v1.0.3"
if [ -n "$1" ]; then
	VERSION=$1
fi

BUILD_DIR="builds"
PACKAGE_NAME="argonaut-${VERSION}"
SOURCE_PATH="."

build_for() {
	os=$1
	arch=$2
	echo "--> Building for ${os}/${arch}..."
	output_name="${BUILD_DIR}/${PACKAGE_NAME}/argonaut-${os}-${arch}"
	if [ "${os}" = "windows" ]; then
	output_name="${output_name}.exe"
	fi
	GOOS=${os} GOARCH=${arch} go build -trimpath -ldflags="-s -w" \
		-o "${output_name}" ${SOURCE_PATH}
}

echo "--> Preparing for release build for version ${VERSION}..."
rm -rf ${BUILD_DIR}
mkdir -p ${BUILD_DIR}/${PACKAGE_NAME}

# --- Tier 1: Core Targets ---
echo "--> Building Core Targets..."
build_for linux amd64
build_for windows amd64
build_for darwin amd64
build_for darwin arm64

# --- Tier 2: Linux with musl-libc
if command -v docker >/dev/null 2>&1; then
	echo "--> Building Linux with musl-libc target..."
	docker build -t argonaut-release -f Dockerfile.release .
	container_id=$(docker create argonaut-release)
	docker cp "${container_id}:/argonaut" \
		"./${BUILD_DIR}/${PACKAGE_NAME}/argonaut-linux-amd64-musl"
	docker rm "${container_id}"
else
	echo "--> WARNING: docker not found. Skipping musl build."
fi


# --- Tier 3: Extended Targets ---
echo "--> Building Extended Targets..."
build_for freebsd amd64
build_for openbsd amd64
build_for netbsd amd64
build_for plan9 amd64

build_for linux riscv64
build_for linux arm64
build_for linux 386

build_for windows arm64
build_for windows 386



# --- Finalization ---
echo "--> Generating SHA256 checksums..."
cd "${BUILD_DIR}/${PACKAGE_NAME}"
sha256sum * > SHA256SUMS.txt
cd ../..

echo ""
echo "--> Done. Artifacts are in '${BUILD_DIR}/${PACKAGE_NAME}'"
