# Security model

## Trust boundaries

The connector handles two short-lived secrets in memory:

1. the one-time gateway pairing capability;
2. the DeepSeek web session token read after an explicit login and a successful
   current-user check inside the temporary browser.

Neither value is written to disk, command output, application logs, crash
reports, HTTP URLs, or the local status page.

On Windows and Linux, the short-lived pairing capability can cross the local
browser/operating-system boundary once through the registered
`chat2api-connector://` launch URI. The URI is not an HTTP request and the
connector does not persist or log it. The DeepSeek token is never placed in the
launch URI.

## Local interface

The setup server:

- binds to `127.0.0.1` on an operating-system-selected port,
- uses a 256-bit random path capability,
- rejects unexpected Host and remote-address values,
- sends no CORS headers,
- enforces a strict Content Security Policy,
- accepts bounded request bodies, and
- exits after success, explicit shutdown, or idle timeout.

## Browser isolation

The connector never attaches to a running browser or its default profile. It
starts a supported installed browser with:

- a newly-created temporary `--user-data-dir`,
- a loopback remote debugging endpoint,
- no extensions,
- no default-browser integration, and
- a hard operation timeout.

Only the exact `https://chat.deepseek.com` origin is eligible for token
extraction. The connector reads only `localStorage.userToken`, accepts the
current versioned JSON envelope's string `value` field or a legacy plain token,
and verifies the normalized value against DeepSeek's current-user endpoint.
Human-verification and sign-in redirects may temporarily expose rejected
placeholder values; these are retried with a bounded backoff and never cause
the isolated login window to close early.
Other storage fields, cookies, passwords, and OTP values are never read. The
profile directory is removed after browser shutdown.

## Gateway submission

Native pairing codes must:

- use the `c2a-ds-native-v1` prefix,
- expire within twelve minutes, including the gateway's ten-minute lifetime
  and bounded clock skew,
- contain a UUID session ID and high-entropy secret,
- target HTTPS, except explicit localhost development,
- use the fixed `/admin/api/deepseek-link/native-complete` path, and
- contain no URL credentials, query, or fragment.

The connector shows the exact gateway host before continuing. HTTP redirects
are rejected so a token cannot be forwarded to another host.

The URL handler is registered only for the current operating-system user and
points to a stable copy under that user's application directory. It never
requires administrator privileges or a background service.

## Residual risks

- DeepSeek can change its login flow or local-storage key.
- Provider anti-automation controls can reject a dedicated browser profile.
- Unsigned development binaries can be blocked or warned about by the OS.
- Local browser or operating-system diagnostics may temporarily observe the
  short-lived pairing launch URI.
- A compromised local machine can inspect any process or browser session.
- A user can approve a malicious gateway hostname without reading it.

Code signing, reproducible release builds, checksums, and a dedicated provider
account are required before broad distribution.
