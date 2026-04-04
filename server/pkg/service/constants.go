package service

import "time"

const machineHeartbeatTTL = 5 * time.Minute
const machineDrainPollInterval = 50 * time.Millisecond
const requestLeaseDuration = 30 * time.Second
const requestDispatchInterval = 2 * time.Second
const requestTimeout = 45 * time.Second
const requestBackoff = 5 * time.Second
const maxMachineConcurrentRequests = 4
