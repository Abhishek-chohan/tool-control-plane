# Compatibility Policy

## Purpose

This document defines how the maintained Toolplane platform evolves without forcing users to infer compatibility rules from commit history.

The authoritative contract source remains `server/proto/service.proto`. `SDK_MAP.md` remains the source of truth for which maintained SDKs expose each RPC as `full`, `partial`, `mock`, or `unsupported`.

## Scope

This policy applies to:

- the protobuf contract in `server/proto/service.proto`
- the maintained HTTP gateway generated from that contract
- the maintained SDK surfaces in Python, Go, and TypeScript

It does not apply to the separate server-side `/rpc` removal path documented in `server/docs/rpc-retirement.md`.

## Protobuf Contract Rules

- Additive protobuf changes are the default compatibility path. New RPCs, new optional fields, and new enum values should preserve existing behavior for unchanged clients.
- Breaking protobuf changes require an explicit version boundary. Removing or renaming RPCs, reusing field numbers, changing wire-visible field types, or changing required request semantics must not ship as an unannounced in-place change on the current major line.
- Deprecated protobuf behavior must be documented before removal. The deprecation target and migration path belong in repo docs, not only in generated code comments.
- The HTTP gateway is coupled to the protobuf contract. If an RPC or field changes in the canonical proto, the generated HTTP gateway behavior changes with it and must be reviewed under the same compatibility policy.

## Maintained SDK Rules

- Python, Go, and TypeScript are the maintained SDK surfaces. Python is the completeness baseline; Go and TypeScript are supported secondary SDKs with narrower public coverage.
- Public wrapper changes must update `SDK_MAP.md` in the same change that lands the code change.
- A maintained SDK may remain `partial` or `unsupported` for an RPC, but that support label must be explicit and the docs must not imply fuller coverage than exists.
- Provider-only and admin-only surfaces may remain Python-only, but they must stay labeled as `provider` or `admin` scope rather than being described as portable SDK guarantees.

## Deprecation Rules

- Public deprecations must be written down before removal. The announcement should identify the affected RPC, wrapper, or behavior, the replacement path, and the planned removal boundary.
- A deprecated surface must remain out of new feature work. Compatibility windows exist for migration, not for expanding legacy ownership.
- Removing a public wrapper requires the code change, the `SDK_MAP.md` update, the relevant docs update, and a release-note entry in the same change set.

## Release Note Expectations

Behavior-changing work must include a release-note-ready summary that answers these questions:

- What changed in the contract or runtime behavior?
- Which services, SDK wrappers, or conformance cases were affected?
- Is the change additive, behavior-tightening, deprecated, or breaking?
- What migration action does an operator, SDK user, or provider author need to take?

When behavior changes affect shared semantics, update the relevant docs and tests in the same change rather than leaving the release note as the only durable explanation.
