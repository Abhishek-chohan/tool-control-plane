package mcp

import (
	"encoding/json"
	"strings"

	gw "toolplane/proto"
)

// taskPollIntervalMs is the polling interval suggested to MCP clients in every
// task payload.
const taskPollIntervalMs = 500

// requestCancelledError is the exact error marker CancelRequest records on a
// cancelled request; it distinguishes user cancellation from other failures,
// which all surface as the generic failure status.
const requestCancelledError = "Request was cancelled"

// taskPayload is the wire shape shared by CreateTaskResult (resultType "task",
// flat Result & Task) and the DetailedTask variants returned by tasks/get
// (resultType "complete"). Field names follow the Tasks extension schema. The
// task identifier is a Toolplane request ID: requests are the durable units
// that provider machines claim and execute.
type taskPayload struct {
	ResultType     string  `json:"resultType"`
	TaskID         string  `json:"taskId"`
	Status         string  `json:"status"`
	StatusMessage  string  `json:"statusMessage,omitempty"`
	CreatedAt      string  `json:"createdAt"`
	LastUpdatedAt  string  `json:"lastUpdatedAt"`
	TTLMs          *int64  `json:"ttlMs"`
	PollIntervalMs int64   `json:"pollIntervalMs,omitempty"`
	Result         any     `json:"result,omitempty"`
	Error          *Error  `json:"error,omitempty"`
	Meta           metaMap `json:"_meta,omitempty"`
}

type metaMap map[string]any

// requestTaskStatus maps a Toolplane request status to an MCP Tasks status.
// pending/claimed/running/stalled all mean work is in flight; MCP has a
// single working state. A failure caused by CancelRequest maps to cancelled;
// every other failure maps to failed.
func requestTaskStatus(request *gw.Request) string {
	switch request.Status {
	case "done":
		return TaskStatusCompleted
	case "failure":
		if request.Error == requestCancelledError {
			return TaskStatusCancelled
		}
		return TaskStatusFailed
	default:
		return TaskStatusWorking
	}
}

// isTerminalStatus reports whether an MCP task status is terminal.
func isTerminalStatus(status string) bool {
	switch status {
	case TaskStatusCompleted, TaskStatusFailed, TaskStatusCancelled:
		return true
	default:
		return false
	}
}

// baseTaskPayload fills the fields every task payload carries. Toolplane
// retains requests indefinitely, so ttlMs is null (unlimited).
func baseTaskPayload(request *gw.Request) taskPayload {
	return taskPayload{
		TaskID:         request.Id,
		Status:         requestTaskStatus(request),
		StatusMessage:  statusMessageForRequest(request),
		CreatedAt:      request.CreatedAt,
		LastUpdatedAt:  request.UpdatedAt,
		TTLMs:          nil,
		PollIntervalMs: taskPollIntervalMs,
	}
}

func statusMessageForRequest(request *gw.Request) string {
	switch request.Status {
	case "pending":
		return "queued for execution"
	case "claimed", "running":
		return "executing"
	case "stalled":
		return "stalled, awaiting lease reclaim"
	case "failure":
		return firstNonEmpty(request.Error, request.ResultType)
	case "cancelled":
		return requestCancelledError
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

// createTaskResult builds the flat CreateTaskResult returned by tools/call
// when the client advertised the Tasks extension.
func createTaskResult(request *gw.Request) taskPayload {
	payload := baseTaskPayload(request)
	payload.ResultType = ResultTypeTask
	return payload
}

// detailedTask builds the tasks/get payload for the request's current status,
// embedding the CallToolResult-shaped final result on completion and a
// JSON-RPC error object on failure, per the extension schema. Cancelled tasks
// carry their reason in statusMessage only (the schema gives CancelledTask no
// error field).
func detailedTask(request *gw.Request) taskPayload {
	payload := baseTaskPayload(request)
	payload.ResultType = ResultTypeComplete
	switch payload.Status {
	case TaskStatusCompleted:
		payload.Result = callToolResultFromRequest(request)
	case TaskStatusFailed:
		payload.Error = &Error{
			Code:    CodeInternalError,
			Message: firstNonEmpty(request.Error, "tool execution failed"),
			Data:    map[string]any{TaskIDDataKey: request.Id},
		}
	}
	return payload
}

// callToolResultFromRequest renders the terminal request outcome as an MCP
// CallToolResult: content blocks plus optional structuredContent, with
// isError set for failures and cancellations.
func callToolResultFromRequest(request *gw.Request) map[string]any {
	result := map[string]any{
		"resultType": ResultTypeComplete,
		"content":    []any{},
	}

	content := []any{}
	if request.Status == "done" {
		if text := strings.TrimSpace(request.Result); text != "" {
			content = append(content, textBlock(text))
		}
		if structured := parseJSONValue(request.Result); structured != nil {
			result["structuredContent"] = structured
		}
	} else {
		message := firstNonEmpty(request.Error, "tool call "+request.Status)
		content = append(content, textBlock("tool call failed: "+message))
		result["isError"] = true
	}
	result["content"] = content
	return result
}

func textBlock(text string) map[string]any {
	return map[string]any{"type": "text", "text": text}
}

// parseJSONValue decodes a JSON document into generic values; returns nil
// when the input is empty or not valid JSON.
func parseJSONValue(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}
	var value any
	if err := json.Unmarshal([]byte(trimmed), &value); err != nil {
		return nil
	}
	return value
}
