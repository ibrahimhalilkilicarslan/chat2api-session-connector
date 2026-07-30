# Native pairing protocol

## Capability

The authenticated Chat2API admin creates an in-memory link session. The native
connector receives a code with this envelope:

```text
c2a-ds-native-v1.<base64url-json>
```

The decoded JSON contains:

```json
{
  "v": 1,
  "transport": "native",
  "endpoint": "https://gateway.example.com/admin/api/deepseek-link/native-complete",
  "sessionId": "uuid",
  "secret": "high-entropy one-time secret",
  "expiresAt": 0
}
```

`expiresAt` is Unix time in milliseconds. The connector accepts at most ten
minutes of remaining lifetime and the gateway currently issues five-minute
sessions.

## Completion request

The connector sends:

```http
POST /admin/api/deepseek-link/native-complete
Content-Type: application/json
X-Chat2API-Connector: native-v1
```

```json
{
  "sessionId": "uuid",
  "secret": "one-time secret",
  "token": "DeepSeek web session token"
}
```

The request has no browser `Origin` header. The gateway rejects requests with an
`Origin`, a missing connector header, the wrong transport secret, an expired
session, or invalid provider credentials.

Native and browser-extension capabilities use different secrets. A capability
issued for one transport cannot complete through the other transport's endpoint.

## Logging

Neither side logs:

- the capability code,
- the one-time secret,
- the provider token,
- request or response bodies containing credentials.

Audit records contain only transport, outcome, provider, account identifier, and
sanitized health code.
