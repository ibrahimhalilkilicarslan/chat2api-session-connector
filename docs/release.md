# Release process

1. Run `make check`.
2. Review `git status` and ensure no secrets or local artifacts are present.
3. Build archives with `make release VERSION=x.y.z`.
4. Run `cd dist && sha256sum --check SHA256SUMS`.
5. Test Windows, macOS, and Linux packages on real machines.
6. Sign Windows binaries with Authenticode.
7. Sign and notarize both macOS application bundles.
8. Publish immutable archives and checksums in a GitHub release.
9. Update the Chat2API gateway connector manifest only after the release exists.

The Linux archives are portable tarballs. A later release may add AppImage,
`.deb`, and `.rpm` packaging without changing the connector protocol.
