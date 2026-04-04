# Machine Pinging And Reconnection Plan

> **Historical implementation note.** This document captures the original design and implementation snapshot for machine pinging and reconnection. For the current authoritative runtime description, see `server/DOCUMENTATION.md`. This file is preserved as design-level reference only.

## Context
- The server acts as a distributed tool-execution control plane over gRPC, coordinating registered machines and tools.
- MCP is a comparison point or possible future adapter target, not the core protocol of this repository.
- Machines register tools and then maintain liveness through periodic ping calls (machine and tool ping updates).
- Cleanup routines remove stale machines and tools to keep the registry accurate for downstream executors and clients.

## Prior Flow
1. **Machine registration**
   - `MachinesService.RegisterMachine` creates or updates a `model.Machine` entry keyed by session and machine identifiers.
   - Incoming tool definitions are re-registered through `ToolService.RegisterTool`; the service tracks tool IDs per machine in `machineTools` (previously keyed only by `machineID`).
2. **Tool re-registration**
   - If a tool name already exists for the session, `ToolService.RegisterTool` reuses the prior `model.Tool`, updates its `MachineID`, and bumps `LastPingAt`.
3. **Heartbeat / ping handling**
   - Machines call `UpdateMachinePing`; tools use `UpdateToolPing` (typically invoked indirectly via reconnect).
   - `cleanupInactiveMachines` runs every five minutes, removing machines whose `LastPingAt` predates the inactivity threshold and deleting tracked tools from `machineTools` using the machine identifier.
4. **Request/task routing**
   - `RequestsService.CreateRequest` validates tool presence via `ToolService.GetToolByName` and locates machines through `MachinesService.FindMachinesWithTool`, which relies on the tool's `MachineID` pointer.

## Observed Gaps
- **Stale machine cleanup removes active tools**: When a tool migrates to a new machine, the old machine retains the tool ID in `machineTools`. During cleanup, those ids trigger `DeleteTool`, wiping the actively re-registered tool.
- **Machine ID namespace collisions**: `machineTools` is keyed solely by machine ID, so the same identifier reused in another session overwrites tracking data and breaks cleanup semantics.
- **Metadata staleness on re-registration**: Re-registering tools do not refresh description, schema, config, or tags; consumers continue to receive outdated metadata.
- **Heartbeat granularity**: Five-minute machine cleanup may be too coarse for an MCP server aiming for fast failover telemetry.

## Proposed Changes
1. **Session-scoped machine tool tracking**
   - Replace `map[string][]string machineTools` with `map[string]map[string][]string` (session → machine → tool IDs).
   - Adjust registration, cleanup, and unregister code paths to reference `(sessionID, machineID)` ensuring multi-tenant safety.
2. **Tool ownership transfer**
   - During tool re-registration, move the tool ID from the previous `(sessionID, machineID)` bucket to the new machine before returning.
   - When cleanup runs, verify that the tool's `MachineID` still matches the stale machine before deleting it.
3. **Refresh tool metadata**
   - Update `ToolService.RegisterTool` to copy incoming description, schema, config, and tags onto the existing `model.Tool` when it reconnects, while preserving created timestamps.
4. **Heartbeat tightening / tooling**
   - Evaluate shorter inactivity windows (e.g., 1–2 minutes) or track `LastPingAt` per tool to catch silent tool failures faster.
   - Consider emitting structured events whenever machines or tools are culled to feed MCP orchestration logic.
5. **Operational instrumentation**
   - Add metrics/log hooks for machine/tool transitions (register, migrate, cleanup) to surface reconnect behavior in observability systems.

## Next Steps
- Prototype the data-structure changes in `MachinesService` and validate cleanup against simulated reconnect flows.
- Add unit tests covering tool migration between machines, concurrent sessions, and stale machine cleanup.
- Revisit gRPC handlers (e.g., streaming updates) once the registry guarantees correctness to ensure MCP clients receive timely state changes.

## Implementation Snapshot
- `MachinesService` now tracks tool ownership per session (`map[sessionID]map[machineID][]toolID`) and migrates registrations when a tool reconnects on a new machine.
- Machine and tool cleanup verifies actual ownership before deletion to avoid stripping active tools during stale-machine eviction.
- `ToolService.RegisterTool` refreshes description, schema, config, and tags when a tool re-registers, preserving the existing record and ID.
