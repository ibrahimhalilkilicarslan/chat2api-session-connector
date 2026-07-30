# Contributing

## Before opening a change

Open an issue for behavior changes that affect the browser boundary, pairing
protocol, credential transfer, or supported platforms. Never attach live
tokens, pairing codes, browser profiles, passwords, OTP values, or private
gateway details.

## Development

Use Go 1.26 or newer and run:

```bash
make check
```

Changes must preserve these invariants:

- only an isolated temporary browser profile is inspected,
- only DeepSeek's `localStorage.userToken` is read,
- provider tokens and pairing capabilities are never persisted or logged,
- gateway submission is restricted to the exact confirmed HTTPS endpoint,
- redirects are rejected,
- local HTTP remains loopback-only,
- no telemetry, background startup, auto-update, or administrator privilege is added.

Add focused tests for platform discovery, protocol parsing, endpoint validation,
redirect rejection, loopback controls, and secret redaction as applicable.
