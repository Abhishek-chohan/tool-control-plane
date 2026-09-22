# Python package: honest versioning and verifiable packaging

Date: 2026-09-21

`toolplane-python-client` metadata no longer claims a release that does
not exist, and the distribution can now be built and inspected from the
repository.

- **The version is derived from the git tag at build time.** The source
  tree previously declared version `1.0.0` with a Production/Stable
  classifier despite the repository having no releases; both claims are
  gone. An untagged tree builds as `0.0.0.dev0`; the release workflow
  injects the tagged version. Maturity is declared as Alpha, matching
  the repository's pre-1.0 compatibility policy.
- **Duplicate/bogus classifiers removed** (a repeated `3.9` and a
  `3.8` below the declared `requires-python >= 3.9` floor).
- **Maintainer contact** now points at the maintainer's GitHub noreply
  address instead of an unused domain.
- **`make build-python`** (from `server/`) builds the wheel and sdist
  into `server/bin/dist` — the same artifacts the release workflow
  publishes — and a new Lint workflow job proves the loop on every
  change: build, `twine check`, wheel install into a clean target, and
  `import toolplane` + `toolplane-provider` entry-point smoke.

Notes: the package contents were already correct (proto stubs,
toolkits, and the `toolplane-provider` console script are all in the
wheel — 35 generated/toolkit files verified); what changed is that the
version is honest and the packaging is CI-proven.
