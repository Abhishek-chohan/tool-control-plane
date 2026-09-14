package service

import "time"

const machineHeartbeatTTL = 5 * time.Minute
const machineDrainPollInterval = 50 * time.Millisecond

// requestLeaseDuration is the lease TTL granted on claim and on each renewal:
// without a renewal the request becomes reclaimable once this window passes.
const requestLeaseDuration = 30 * time.Second
const requestDispatchInterval = 2 * time.Second

// requestTimeout is the default absolute per-attempt execution timeout. It
// bounds the total runtime of one attempt regardless of lease renewals;
// callers may override it per request via CreateRequest/ExecuteTool
// timeout_seconds.
const requestTimeout = 45 * time.Second

// maxRequestTimeout caps caller-supplied per-request timeout overrides.
const maxRequestTimeout = time.Hour
const requestBackoff = 5 * time.Second
const maxMachineConcurrentRequests = 4

// waitPollInterval is the fallback poll cadence for non-claiming waiters
// (tasks, synchronous execution) — request-update signals are the primary
// wake-up; the ticker covers completions made on another replica.
const waitPollInterval = 250 * time.Millisecond

// requestRetentionAge bounds how long terminal requests are retained;
// requestCleanupInterval is the retention sweeper cadence.
const requestRetentionAge = 7 * 24 * time.Hour
const requestCleanupInterval = 30 * time.Minute

// auditRetentionAge bounds how long audit events are retained.
const auditRetentionAge = 30 * 24 * time.Hour

// maxPendingRequestsPerSession caps the outstanding pending backlog a single
// session may accumulate. Providers are expected to drain it; a session
// whose claims stall (no provider, capacity) hits this ceiling instead of
// growing the queue without bound.
const maxPendingRequestsPerSession = 512
