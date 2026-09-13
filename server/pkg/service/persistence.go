package service

import "time"

const defaultPersistenceTimeout = 2 * time.Second

// startupLoadTimeout bounds construction-time cache hydration (tools,
// machines, sessions). These loads run once at startup and compete with every
// other workload on the database; a short per-call bound made them fail under
// load, leaving caches silently empty.
const startupLoadTimeout = 30 * time.Second
