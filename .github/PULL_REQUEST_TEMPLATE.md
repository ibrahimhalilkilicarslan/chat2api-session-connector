## Summary

Describe the problem and the smallest cross-platform change that solves it.

## Security boundary

- [ ] The connector still uses a new temporary browser profile.
- [ ] Only DeepSeek's `localStorage.userToken` is inspected.
- [ ] Tokens and pairing capabilities are not persisted or logged.
- [ ] Gateway redirects remain rejected.
- [ ] Local HTTP remains loopback-only.
- [ ] No telemetry, background service, or administrator privilege was added.

## Platform verification

- [ ] Windows behavior considered
- [ ] macOS behavior considered
- [ ] Linux behavior considered
- [ ] `make check`

## Evidence

List safe tests and diagnostics. Never include live tokens, pairing codes,
browser profiles, passwords, OTP values, or private endpoint details.
