package model

import (
	"time"

	"github.com/google/uuid"
)

// RequestStatus represents the status of a request
type RequestStatus string

const (
	RequestStatusPending RequestStatus = "pending"
	RequestStatusClaimed RequestStatus = "claimed"
	RequestStatusRunning RequestStatus = "running"
	RequestStatusDone    RequestStatus = "done"
	RequestStatusFailed  RequestStatus = "failure"
	RequestStatusStalled RequestStatus = "stalled"
)

// ResultType represents the type of result
type ResultType string

const (
	ResultTypeResolution ResultType = "resolution"
	ResultTypeRejection  ResultType = "rejection"
	ResultTypeInterrupt  ResultType = "interrupt"
	ResultTypeStreaming  ResultType = "streaming"
)

// Request represents a tool execution request
type Request struct {
	ID                 string            `json:"id"`
	SessionID          string            `json:"sessionId"`
	ToolName           string            `json:"toolName"`
	Status             RequestStatus     `json:"status"`
	Input              string            `json:"input"`
	Result             interface{}       `json:"result"`
	ResultType         ResultType        `json:"resultType,omitempty"`
	StreamResults      []string          `json:"streamResults,omitempty"`
	StreamStartSeq     int32             `json:"streamStartSeq"`
	NextStreamSeq      int32             `json:"nextStreamSeq"`
	Error              string            `json:"error,omitempty"`
	Meta               map[string]string `json:"meta,omitempty"`
	ExecutingMachineID string            `json:"executingMachineId,omitempty"`
	Attempts           int               `json:"attempts"`
	MaxAttempts        int               `json:"maxAttempts"`
	TimeoutSeconds     int               `json:"timeoutSeconds"`
	BackoffSeconds     int               `json:"backoffSeconds"`
	VisibleAt          time.Time         `json:"visibleAt"`
	NextAttemptAt      *time.Time        `json:"nextAttemptAt,omitempty"`
	LeasedBy           string            `json:"leasedBy,omitempty"`
	LeasedAt           *time.Time        `json:"leasedAt,omitempty"`
	// LeaseEpoch identifies the current claim grant. It is incremented on every
	// successful claim and must be echoed by the lease holder on fenced writes
	// (update/submit/append/renew). It is retained after the request reaches a
	// terminal state so late writes from a stale executor can still be rejected.
	LeaseEpoch int64 `json:"leaseEpoch"`
	// IdempotencyKey, when set by the caller, dedups creates within the
	// session: re-creating with the same key returns the original request.
	IdempotencyKey string    `json:"idempotencyKey,omitempty"`
	LastError      string    `json:"lastError,omitempty"`
	DeadLetter     bool      `json:"deadLetter"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// NewRequest creates a new request
func NewRequest(sessionID, toolName, input string) *Request {
	now := time.Now()
	return &Request{
		ID:             uuid.New().String(),
		SessionID:      sessionID,
		ToolName:       toolName,
		Status:         RequestStatusPending,
		Input:          input,
		CreatedAt:      now,
		UpdatedAt:      now,
		StreamResults:  []string{},
		StreamStartSeq: 1,
		NextStreamSeq:  1,
		Meta:           make(map[string]string),
		Attempts:       0,
		MaxAttempts:    defaultRequestMaxAttempts,
		TimeoutSeconds: defaultRequestTimeoutSeconds,
		BackoffSeconds: defaultRequestBackoffSeconds,
		VisibleAt:      now,
	}
}

// SetRunning sets the request status to running
func (r *Request) SetRunning(machineID string) {
	r.Status = RequestStatusRunning
	r.ExecutingMachineID = machineID
	r.UpdatedAt = time.Now()
}

// SetClaimedBy marks the request as claimed by a machine
func (r *Request) SetClaimedBy(machineID string) {
	r.Status = RequestStatusClaimed
	r.ExecutingMachineID = machineID
	r.UpdatedAt = time.Now()
}

// SetResult sets the request result and marks as complete
func (r *Request) SetResult(result interface{}, resultType ResultType, errorMsg string) {
	r.Result = result
	r.ResultType = resultType
	r.Error = errorMsg

	if resultType == ResultTypeRejection {
		r.Status = RequestStatusFailed
	} else {
		r.Status = RequestStatusDone
	}

	r.UpdatedAt = time.Now()
	r.LeasedBy = ""
	r.LeasedAt = nil
	r.VisibleAt = time.Now()
	r.NextAttemptAt = nil
}

// RequestChunkWindow describes the retained stream chunk window.
type RequestChunkWindow struct {
	StartSeq int32
	NextSeq  int32
	Chunks   []string
}

// Clone returns a copy safe to hand out of the service's cache: the struct
// is copied field-by-field, and the mutable reference fields (Meta,
// StreamResults, the time pointers) get fresh copies so callers reading the
// clone never race with the service mutating the cached original. Result is
// shared by reference: service code only ever assigns it wholesale, never
// mutates it in depth.
func (r *Request) Clone() *Request {
	cloned := *r
	if r.Meta != nil {
		cloned.Meta = make(map[string]string, len(r.Meta))
		for k, v := range r.Meta {
			cloned.Meta[k] = v
		}
	}
	if r.StreamResults != nil {
		cloned.StreamResults = make([]string, len(r.StreamResults))
		copy(cloned.StreamResults, r.StreamResults)
	}
	if r.NextAttemptAt != nil {
		t := *r.NextAttemptAt
		cloned.NextAttemptAt = &t
	}
	if r.LeasedAt != nil {
		t := *r.LeasedAt
		cloned.LeasedAt = &t
	}
	return &cloned
}

// EnsureStreamSequenceDefaults normalizes retained-window sequence state for legacy requests.
func (r *Request) EnsureStreamSequenceDefaults() {
	if r.StreamStartSeq <= 0 {
		r.StreamStartSeq = 1
	}
	minimumNextSeq := r.StreamStartSeq + int32(len(r.StreamResults))
	if r.NextStreamSeq < minimumNextSeq {
		r.NextStreamSeq = minimumNextSeq
	}
}

// AddStreamChunk adds a chunk to the retained stream window and returns its absolute sequence number.
// The window is bounded twice: by count (maxRequestStreamChunks) and by total
// payload bytes (maxRequestStreamWindowBytes). When a bound is exceeded the
// oldest chunks are dropped and StartSeq advances, so consumers resuming from
// a trimmed sequence see the retained window's true start.
func (r *Request) AddStreamChunk(chunk string) int32 {
	r.EnsureStreamSequenceDefaults()
	seq := r.NextStreamSeq
	r.StreamResults = append(r.StreamResults, chunk)
	r.NextStreamSeq++
	r.trimStreamWindow()
	r.UpdatedAt = time.Now()
	return seq
}

// trimStreamWindow drops oldest chunks until the retained window fits both
// the count and byte bounds, advancing StartSeq to match.
func (r *Request) trimStreamWindow() {
	// Find the retained prefix in one backward pass: the newest chunks that
	// fit both the count and byte bounds. Copying the remainder into a fresh
	// slice releases the dropped chunks' backing array for GC.
	total := 0
	keep := len(r.StreamResults)
	for i := len(r.StreamResults) - 1; i >= 0; i-- {
		total += len(r.StreamResults[i])
		if total > MaxRequestStreamWindowBytes || (len(r.StreamResults)-i) > maxRequestStreamChunks {
			// chunk i is the oldest one that no longer fits: everything after it
			// is the retained window.
			keep = len(r.StreamResults) - i - 1
			break
		}
	}
	drop := len(r.StreamResults) - keep
	if drop <= 0 {
		return
	}
	r.StreamStartSeq += int32(drop)
	if keep == 0 {
		r.StreamResults = nil
		return
	}
	retained := make([]string, keep)
	copy(retained, r.StreamResults[len(r.StreamResults)-keep:])
	r.StreamResults = retained
}

func (r *Request) ClearStreamChunks() {
	r.EnsureStreamSequenceDefaults()
	r.StreamResults = []string{}
	r.StreamStartSeq = r.NextStreamSeq
	r.UpdatedAt = time.Now()
}

// StreamChunkWindow returns a copy of the current retained chunk window.
func (r *Request) StreamChunkWindow() RequestChunkWindow {
	r.EnsureStreamSequenceDefaults()
	chunks := append([]string(nil), r.StreamResults...)
	return RequestChunkWindow{
		StartSeq: r.StreamStartSeq,
		NextSeq:  r.NextStreamSeq,
		Chunks:   chunks,
	}
}

// StreamChunkWindowAfter returns the retained chunk window beginning after lastSeq.
func (r *Request) StreamChunkWindowAfter(lastSeq int32) RequestChunkWindow {
	window := r.StreamChunkWindow()
	if lastSeq < 0 {
		lastSeq = 0
	}
	nextSeq := lastSeq + 1
	if nextSeq < window.StartSeq {
		nextSeq = window.StartSeq
	}
	if nextSeq >= window.NextSeq {
		return RequestChunkWindow{StartSeq: window.NextSeq, NextSeq: window.NextSeq, Chunks: []string{}}
	}
	startIndex := int(nextSeq - window.StartSeq)
	return RequestChunkWindow{
		StartSeq: nextSeq,
		NextSeq:  window.NextSeq,
		Chunks:   append([]string(nil), window.Chunks[startIndex:]...),
	}
}

// FinalStreamSeq returns the reserved sequence number for the terminal callback marker.
func (r *Request) FinalStreamSeq() int32 {
	r.EnsureStreamSequenceDefaults()
	return r.NextStreamSeq
}

const (
	defaultRequestMaxAttempts    = 3
	defaultRequestTimeoutSeconds = 45
	defaultRequestBackoffSeconds = 5
	maxRequestStreamChunks       = 100
	// MaxRequestStreamWindowChunks is the stored-window count bound, shared
	// with the chunk-table trim (storage layer cannot see unexported consts).
	MaxRequestStreamWindowChunks = 100
	// MaxRequestChunkBytes bounds one stream chunk; larger payloads are
	// rejected at the RPC boundary rather than truncated mid-stream.
	MaxRequestChunkBytes = 512 << 10
	// MaxChunkBatchBytes bounds one AppendRequestChunks RPC: the server's
	// MaxRecvMsgSize is set to match, so a full max-size batch fits.
	MaxChunkBatchBytes = 32 * MaxRequestChunkBytes
	// maxRequestStreamWindowBytes bounds the retained window: when the
	// window's total payload exceeds it, oldest chunks are trimmed (StartSeq
	// advances) until it fits.
	// MaxRequestStreamWindowBytes is the stored-window byte bound (shared
	// with the chunk-table trim).
	MaxRequestStreamWindowBytes = 8 << 20
)

// ScheduleRetry computes the next attempt time with linear backoff.
func (r *Request) ScheduleRetry() time.Time {
	r.Attempts++
	backoff := time.Duration(r.BackoffSeconds*r.Attempts) * time.Second
	next := time.Now().Add(backoff)
	r.NextAttemptAt = &next
	r.VisibleAt = next
	r.UpdatedAt = time.Now()
	return next
}

// MarkDeadLetter marks the request as terminal after exhausting retries.
func (r *Request) MarkDeadLetter(errMsg string) {
	r.DeadLetter = true
	r.Status = RequestStatusFailed
	r.LastError = errMsg
	r.UpdatedAt = time.Now()
}

// HasTimedOut returns true when the request's absolute per-attempt deadline
// (leased_at + timeout_seconds) has passed. Renewing the lease does not move
// this deadline; it bounds how long one attempt may run regardless of renewals.
func (r *Request) HasTimedOut(now time.Time) bool {
	if r.LeasedAt == nil {
		return false
	}
	deadline := r.LeasedAt.Add(time.Duration(r.TimeoutSeconds) * time.Second)
	return now.After(deadline)
}

// LeaseExpired returns true when the current lease deadline (visible_at) has
// passed without a renewal. It only applies to leased requests; callers should
// only consult it for claimed/running requests.
func (r *Request) LeaseExpired(now time.Time) bool {
	if r.LeasedAt == nil {
		return false
	}
	return !r.VisibleAt.IsZero() && !r.VisibleAt.After(now)
}
