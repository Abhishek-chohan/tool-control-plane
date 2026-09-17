package main

import (
	"strings"

	proto "toolplane/proto"
)

// statusFromName maps a user-facing status name onto the wire enum value.
// Unknown names produce the zero enum, which the server treats as "no
// filter" for list requests.
func statusFromName(name string) proto.RequestStatus {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "pending":
		return proto.RequestStatus_REQUEST_STATUS_PENDING
	case "claimed":
		return proto.RequestStatus_REQUEST_STATUS_CLAIMED
	case "running":
		return proto.RequestStatus_REQUEST_STATUS_RUNNING
	case "done":
		return proto.RequestStatus_REQUEST_STATUS_DONE
	case "failed":
		return proto.RequestStatus_REQUEST_STATUS_FAILED
	case "cancelled":
		return proto.RequestStatus_REQUEST_STATUS_CANCELLED
	default:
		return 0
	}
}

// stringsTrimPrefix removes a prefix from s.
func stringsTrimPrefix(s, prefix string) string {
	return strings.TrimPrefix(s, prefix)
}
