# Chat2API Session Connector

## Mission

Build a small, cross-platform local helper that links a user-authorized DeepSeek
web session to a trusted Chat2API gateway. The connector must use an installed
Chromium browser and must not bundle Electron, a browser runtime, or provider
credentials.

## Security invariants

- Never inspect a user's default browser profile.
- Launch only a dedicated, non-default temporary browser profile.
- Bind local HTTP services only to loopback with an unguessable session path.
- Read only `localStorage.userToken` and only from `https://chat.deepseek.com`.
- Never read passwords, OTP values, cookies, browsing history, or unrelated
  local storage.
- Never persist or log the DeepSeek token or pairing capability.
- Accept only short-lived `c2a-ds-native-v1` capabilities.
- Send credentials only to the exact validated HTTPS gateway endpoint shown to
  and confirmed by the user.
- Reject redirects when submitting credentials.
- Keep Windows, macOS, and Linux behavior equivalent.
- Do not add auto-update, background startup, telemetry, or analytics.

## Commands

```bash
make fmt
make vet
make test
make build
make check
make release
```

## Change policy

- Keep platform-specific behavior behind small files selected by build tags.
- Add tests for capability parsing, endpoint restrictions, redirect rejection,
  loopback UI controls, browser discovery, and secret redaction.
- Never claim a browser/provider smoke passed unless it ran on that platform.
- Document that DeepSeek's web session protocol is unofficial and may change.
