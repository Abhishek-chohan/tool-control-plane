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
	LastError          string            `json:"lastError,omitempty"`
	DeadLetter         bool              `json:"deadLetter"`
	CreatedAt          time.Time         `json:"createdAt"`
	UpdatedAt          time.Time         `json:"updatedAt"`
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
func (r *Request) AddStreamChunk(chunk string) int32 {
	r.EnsureStreamSequenceDefaults()
	seq := r.NextStreamSeq
	r.StreamResults = append(r.StreamResults, chunk)
	r.NextStreamSeq++
	if len(r.StreamResults) > maxRequestStreamChunks {
		over := len(r.StreamResults) - maxRequestStreamChunks
		r.StreamResults = append([]string(nil), r.StreamResults[over:]...)
		r.StreamStartSeq += int32(over)
	}
	r.UpdatedAt = time.Now()
	return seq
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

// HasTimedOut returns true when the request lease timed out.
func (r *Request) HasTimedOut(now time.Time) bool {
	if r.LeasedAt == nil {
		return false
	}
	deadline := r.LeasedAt.Add(time.Duration(r.TimeoutSeconds) * time.Second)
	return now.After(deadline)
}
