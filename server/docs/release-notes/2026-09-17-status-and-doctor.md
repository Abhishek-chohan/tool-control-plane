# toolplane status and doctor

Date: 2026-09-17

Two operator commands complete the onboarding and day-2 story:

- **`toolplane status`**: one screen of truth — reachability, the
  server's build version, and its resolved storage mode (memory or
  postgres, from the HealthCheck response). Unreachable servers print
  the `toolplane serve` next step.
- **`toolplane doctor`**: validates configuration (the environment
  contract and production gates, without starting a server),
  connectivity (with the serve hint on failure), and version skew
  (client binary vs server build). With `--database-url` it also checks
  database reachability and schema freshness — `OpenFromEnv` runs
  migrations on connect, so a successful open is the check. Every line
  names what passed, what failed, and the fix.

- **Additive wire field**: `HealthCheckResponse.storage` (string) now
  carries the resolved storage mode; servers predating the field return
  an empty value, which the CLI renders as "predates storage
  reporting". Per the compatibility policy this is additive and safe
  for old clients.

Notes: `doctor`'s config check runs against the calling process's
environment — it is a pre-flight for `toolplane serve` on the same
host. Version skew needs both sides to report a version (T4's
ldflags-injected builds; source builds report `dev` on both sides and
compare equal).
