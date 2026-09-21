# TypeScript package: packable manifest

Date: 2026-09-21

`toolplane-typescript-client` is now a well-formed publishable npm
package rather than a dev checkout wearing a version number.

- **Version `1.0.0` → `0.0.0-dev`.** The old value claimed a release
  that does not exist; the release workflow stamps the tagged version
  at publish time (`npm version`), and `prepublishOnly` guarantees a
  fresh build.
- **`files` allowlist**: the published tarball contains exactly
  `dist/` (compiled client plus the copied proto assets), the README,
  and the license — 49 files, verified; `src/`, `tests/`, and build
  scripts no longer leak into the artifact.
- **Repository/bugs/homepage/engines metadata** added (`node >= 18`,
  matching the toolchain), so the package page on npm is complete and
  `npm install` warns on unsupported Node versions.

Notes: verified locally via the same loop the Lint workflow's
`packaging` job runs — build, `npm pack`, tarball content assertions,
and an install-and-`require` smoke into a clean project. The package
still is not published; this change only makes the artifact correct
when it is.
