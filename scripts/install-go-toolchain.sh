#!/usr/bin/env bash
set -Eeuo pipefail

VERSION="go1.26.5"
ARCHIVE="${VERSION}.linux-amd64.tar.gz"
SHA256="5c2c3b16caefa1d968a94c1daca04a7ca301a496d9b086e17ad77bb81393f053"
DESTINATION="${HOME}/.local/share/go/${VERSION}"
DOWNLOAD="${TMPDIR:-/tmp}/${ARCHIVE}"

if [[ -x "${DESTINATION}/bin/go" ]]; then
  "${DESTINATION}/bin/go" version
  exit 0
fi

mkdir -p -- "${DESTINATION}"
curl -fL --retry 3 --proto '=https' --tlsv1.2 \
  "https://go.dev/dl/${ARCHIVE}" \
  -o "${DOWNLOAD}"
printf '%s  %s\n' "${SHA256}" "${DOWNLOAD}" | sha256sum --check --strict
tar -xzf "${DOWNLOAD}" --strip-components=1 -C "${DESTINATION}"
"${DESTINATION}/bin/go" version
