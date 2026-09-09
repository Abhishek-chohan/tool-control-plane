package service

import (
	"toolplane/pkg/model"
)

// mutateCachedRequestForTest applies fn to the service's cached copy of the
// request under the write lock. Tests use it to forge lease deadlines and
// visibility times now that the read APIs hand out clones instead of the
// cached pointers.
func mutateCachedRequestForTest(s *RequestsService, requestID string, fn func(*model.Request)) bool {
	s.requestsMutex.Lock()
	defer s.requestsMutex.Unlock()
	for _, sessionRequests := range s.requests {
		if req, ok := sessionRequests[requestID]; ok {
			fn(req)
			return true
		}
	}
	return false
}
