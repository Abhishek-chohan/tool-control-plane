# toolplane serve --config: YAML as the base configuration layer

Date: 2026-09-17

`toolplane serve` accepts a YAML config file that supplies the base
layer of its configuration. Precedence is fixed and documented:
**flags override environment variables, which override the file, which
overrides defaults.** The file feeds the same environment contract the
binaries already use, so the production gates apply to file-sourced
values with no second validation path.

```yaml
env: production
auth:
  mode: postgres
storage:
  mode: postgres
  database_url: ${DATABASE_URL}      # ${VAR} expansion everywhere
  # database_url_file: /run/secrets/db_url   # secret indirection
server:
  port: 9001
  metrics_listen: ":9102"
  cert_file: /etc/toolplane/server.crt
  key_file: /etc/toolplane/server.key
```

- **Unknown keys are a boot failure naming the key** — a renamed key
  must never silently ignore a value the operator believes is live.
- **Secrets never have to be literals**: values expand `${VAR}`
  references, and `storage.database_url_file` reads the database URL
  from a mounted secret file. A group/world-readable config file logs a
  loud warning since it can carry secrets.
- **`--dry-run`** prints the fully resolved configuration with per-key
  provenance — `(flag)`, `(env)`, `(file)`, `(default)` — and exits
  without serving. Secrets render as `***set***`, never in cleartext.
  This is the first stop for "why did it boot this way?"

Notes: the file's environment-backed keys apply only when the variable
is unset, which is what makes the precedence fall out of the ordinary
resolution order rather than a parallel config system. `server.port`
and `server.metrics_listen` have no environment variable and are
applied by the entry point when the corresponding flag was not passed.
