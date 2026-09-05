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
