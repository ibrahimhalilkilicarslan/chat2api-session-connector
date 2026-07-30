# Chat2API Session Connector

Cross-platform local helper for linking a user-authorized DeepSeek web session
to a Chat2API gateway without a browser extension.

The connector:

- opens a local loopback-only confirmation page,
- validates a ten-minute one-time pairing capability,
- launches an installed Chrome, Edge, Chromium, or Brave browser with an
  isolated temporary profile,
- waits for the operator to sign in directly on DeepSeek,
- reads only DeepSeek's `localStorage.userToken`, normalizes DeepSeek's current
  versioned `{ "value": "..." }` envelope with legacy plain-token
  compatibility, and verifies it against DeepSeek's current-user endpoint,
- posts the token to the exact confirmed gateway endpoint, and
- destroys the temporary browser profile when the operation ends.

It does not bundle a browser, read the default browser profile, capture
passwords or OTP values, persist tokens, or run in the background.

> [!IMPORTANT]
> DeepSeek does not document this web-session flow as a public API. Provider
> changes, account controls, or terms may interrupt it. Use a dedicated account
> and review the applicable provider terms.

## Development

Requirements: Go 1.26 or newer.

```bash
make check
```

Run locally:

```bash
go run ./cmd/chat2api-connector
```

On Windows and Linux, the first launch installs a per-user
`chat2api-connector://` protocol handler. Chat2API can then open the connector
with the one-time pairing capability already loaded. The loopback page still
shows the exact gateway hostname before any provider session is transferred.

Launching the connector directly after installation shows a short completion
guide and sends the user back to Chat2API instead of asking for a pairing code.
Manual code entry remains available as a recovery path and is the current
onboarding method on macOS.

On Windows, the connector prefers the supported browser configured as the
system default. If that browser is unsupported, it falls back to an installed
Chrome, Edge, Brave, or Chromium executable without reading the user's regular
browser profile.

## Releases

```bash
make release VERSION=0.2.0
```

Release archives are written to `dist/` for:

- Windows amd64 and arm64
- macOS amd64 and arm64
- Linux amd64 and arm64

Unsigned development builds may trigger Windows SmartScreen or macOS Gatekeeper.
Public distribution should use Authenticode signing and Apple Developer ID
signing/notarization.

See [Security](docs/security.md), [Protocol](docs/protocol.md), and
[Release](docs/release.md).

## License

MIT
