#!/usr/bin/env bash
set -Eeuo pipefail

ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION="${VERSION:-0.2.0}"
COMMIT="${COMMIT:-$(git -C "${ROOT}" rev-parse --short=12 HEAD 2>/dev/null || printf 'unknown')}"
DIST="${ROOT}/dist"
MODULE="github.com/ibrahimhalilkilicarslan/chat2api-session-connector/internal/version"
BASE_LDFLAGS="-s -w -buildid= -X ${MODULE}.Version=${VERSION} -X ${MODULE}.Commit=${COMMIT}"

rm -rf -- "${DIST}"
mkdir -p -- "${DIST}/work"

build_binary() {
  local os="$1"
  local arch="$2"
  local output="$3"
  local extra_ldflags="${4:-}"
  CGO_ENABLED=0 GOOS="${os}" GOARCH="${arch}" \
    go build -trimpath -ldflags "${BASE_LDFLAGS} ${extra_ldflags}" \
    -o "${output}" "${ROOT}/cmd/chat2api-connector"
}

package_windows() {
  local arch="$1"
  local work="${DIST}/work/windows-${arch}"
  mkdir -p -- "${work}"
  build_binary windows "${arch}" "${work}/Chat2API-Connector.exe" "-H=windowsgui"
  mkdir -p -- "${work}/docs"
  cp -- "${ROOT}/LICENSE" "${ROOT}/README.md" "${work}/"
  cp -- "${ROOT}/docs/protocol.md" "${ROOT}/docs/security.md" "${work}/docs/"
  (
    cd -- "${work}"
    zip -q -9 -r "${DIST}/chat2api-session-connector_${VERSION}_windows_${arch}.zip" ./*
  )
}

package_linux() {
  local arch="$1"
  local work="${DIST}/work/linux-${arch}"
  mkdir -p -- "${work}"
  build_binary linux "${arch}" "${work}/chat2api-connector"
  chmod 0755 "${work}/chat2api-connector"
  mkdir -p -- "${work}/docs"
  cp -- "${ROOT}/LICENSE" "${ROOT}/README.md" "${work}/"
  cp -- "${ROOT}/docs/protocol.md" "${ROOT}/docs/security.md" "${work}/docs/"
  tar -C "${work}" -czf "${DIST}/chat2api-session-connector_${VERSION}_linux_${arch}.tar.gz" .
}

package_macos() {
  local arch="$1"
  local work="${DIST}/work/darwin-${arch}"
  local app="${work}/Chat2API Connector.app"
  mkdir -p -- "${app}/Contents/MacOS"
  build_binary darwin "${arch}" "${app}/Contents/MacOS/chat2api-connector"
  chmod 0755 "${app}/Contents/MacOS/chat2api-connector"
  cat > "${app}/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleDisplayName</key><string>Chat2API Connector</string>
  <key>CFBundleExecutable</key><string>chat2api-connector</string>
  <key>CFBundleIdentifier</key><string>com.chat2api.session-connector</string>
  <key>CFBundleInfoDictionaryVersion</key><string>6.0</string>
  <key>CFBundleName</key><string>Chat2API Connector</string>
  <key>CFBundlePackageType</key><string>APPL</string>
  <key>CFBundleShortVersionString</key><string>${VERSION}</string>
  <key>CFBundleVersion</key><string>${VERSION}</string>
  <key>LSMinimumSystemVersion</key><string>12.0</string>
  <key>LSUIElement</key><true/>
  <key>NSHighResolutionCapable</key><true/>
</dict>
</plist>
PLIST
  mkdir -p -- "${work}/docs"
  cp -- "${ROOT}/LICENSE" "${ROOT}/README.md" "${work}/"
  cp -- "${ROOT}/docs/protocol.md" "${ROOT}/docs/security.md" "${work}/docs/"
  (
    cd -- "${work}"
    zip -q -9 -r "${DIST}/chat2api-session-connector_${VERSION}_macos_${arch}.zip" .
  )
}

package_windows amd64
package_windows arm64
package_macos amd64
package_macos arm64
package_linux amd64
package_linux arm64

(
  cd -- "${DIST}"
  sha256sum ./*.zip ./*.tar.gz > SHA256SUMS
)
rm -rf -- "${DIST}/work"
printf 'Release artifacts created in %s\n' "${DIST}"
