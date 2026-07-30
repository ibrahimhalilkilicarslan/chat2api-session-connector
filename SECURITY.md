# Security Policy

## Reporting

Do not open a public issue containing provider tokens, pairing capabilities,
browser profile data, gateway credentials, passwords, OTP values, or private
endpoint details.

Report vulnerabilities privately through
[GitHub Security Advisories](https://github.com/ibrahimhalilkilicarslan/chat2api-session-connector/security/advisories/new).
Include the affected version, impact, safe reproduction steps, and a proposed
fix when available.

## Supported versions

Only the latest published version is intended to receive security fixes.

## Trust boundary

The connector is designed to read one explicitly authorized DeepSeek web
session from an isolated temporary browser profile and submit it once to an
operator-confirmed HTTPS gateway. It must not read a default browser profile,
persist secrets, follow gateway redirects, capture login fields, or run as a
background service.

The complete design and residual risks are documented in
[docs/security.md](docs/security.md).
