package mcp

import (
	"encoding/json"
	"strings"

	gw "toolplane/proto"
)

// taskPollIntervalMs is the polling interval suggested to MCP clients in
// every task payload.
const taskPollIntervalMs = 500

// taskPayload is the wire shape shared by CreateTaskResult (resultType "task",
// flat Result & Task) and the DetailedTask variants returned by tasks/get
// (resultType "complete"). Field names follow the Tasks extension schema.
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

// mapTaskStatus translates a Toolplane task status into an MCP Tasks status.
// Toolplane's pending and running states both mean work is in flight; MCP has
// a single working state. Toolplane never reports input_required.
func mapTaskStatus(status string) string {
	switch status {
	case "completed":
		return TaskStatusCompleted
	case "failed":
		return TaskStatusFailed
	case "cancelled":
		return TaskStatusCancelled
	default: // pending, running, and any future in-flight state
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
// retains tasks indefinitely, so ttlMs is null (unlimited).
func baseTaskPayload(task *gw.Task) taskPayload {
	return taskPayload{
		TaskID:         task.Id,
		Status:         mapTaskStatus(task.Status),
		StatusMessage:  statusMessageFor(task),
		CreatedAt:      task.CreatedAt,
		LastUpdatedAt:  task.UpdatedAt,
		TTLMs:          nil,
		PollIntervalMs: taskPollIntervalMs,
	}
}

func statusMessageFor(task *gw.Task) string {
	switch task.Status {
	case "pending":
		return "queued for execution"
	case "running":
		if task.CurrentRequestId != "" {
			return "executing"
		}
		return "running"
	case "failed", "cancelled":
		return firstNonEmpty(task.Error, task.ResultType)
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
func createTaskResult(task *gw.Task) taskPayload {
	payload := baseTaskPayload(task)
	payload.ResultType = ResultTypeTask
	return payload
}

// detailedTask builds the tasks/get payload for the task's current status,
// embedding the CallToolResult-shaped final result on completion and a
// JSON-RPC error object on failure, per the extension schema.
func detailedTask(task *gw.Task) taskPayload {
	payload := baseTaskPayload(task)
	payload.ResultType = ResultTypeComplete
	switch payload.Status {
	case TaskStatusCompleted:
		payload.Result = callToolResultFromTask(task)
	case TaskStatusFailed:
		payload.Error = &Error{
			Code:    CodeInternalError,
			Message: firstNonEmpty(task.Error, "task failed"),
			Data:    map[string]any{TaskIDDataKey: task.Id},
		}
	}
	return payload
}

// callToolResultFromTask renders the terminal task outcome as an MCP
// CallToolResult: content blocks plus optional structuredContent, with
// isError set for failures and cancellations.
func callToolResultFromTask(task *gw.Task) map[string]any {
	result := map[string]any{
		"resultType": ResultTypeComplete,
		"content":    []any{},
	}

	content := []any{}
	if task.Status == "completed" {
		if text := strings.TrimSpace(task.Result); text != "" {
			content = append(content, textBlock(text))
		}
		if structured := parseJSONValue(task.Result); structured != nil {
			result["structuredContent"] = structured
		}
	} else {
		message := firstNonEmpty(task.Error, "task "+task.Status)
		content = append(content, textBlock("tool call "+task.Status+": "+message))
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
