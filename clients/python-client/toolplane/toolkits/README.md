# Toolplane Toolkits

Optional tool collections for Toolplane providers. Toolkits are opt-in:
the Python SDK imports nothing from here by default.

## `standalone_tools/`

Filesystem, search, and test-analysis tools (create/read/edit files,
grep/glob search, semantic search, test-failure analysis) usable as a
plain CLI or registered as provider tools. See
[standalone_tools/README.md](standalone_tools/README.md) for the catalog,
the `TOOLPLANE_WORKSPACE_ROOT` / `TOOLPLANE_BASH_TIMEOUT_SECONDS`
guardrails, and how to run the test suite.

> **⚠️ UNSANDBOXED.** These tools operate with the full privileges of
> the running process. Register them on a provider only when the provider
> itself runs inside an isolated container or VM — see the security
> guidance in `server/docs/release-notes/` (security posture note).
