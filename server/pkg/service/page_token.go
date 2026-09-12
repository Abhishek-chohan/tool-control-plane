package service

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// pageTokenCodec converts numeric list offsets into opaque v1 page tokens.
// The token is deliberately non-enumerable but carries no secrets: it is a
// base64-wrapped offset that clients pass back verbatim. Changing the wire
// format invalidates outstanding tokens, which is acceptable — clients
// simply restart from the first page.
var pageTokenCodec pageTokenCodecType

type pageTokenCodecType struct{}

// Encode renders an offset as an opaque token.
func (pageTokenCodecType) Encode(offset int) (string, error) {
	if offset < 0 {
		return "", fmt.Errorf("negative offset %d", offset)
	}
	if offset == 0 {
		return "", nil
	}
	return base64.RawURLEncoding.EncodeToString([]byte("offset:" + strconv.Itoa(offset))), nil
}

// Decode parses a token into its offset. An empty token decodes to 0.
func (pageTokenCodecType) Decode(token string) (int, error) {
	if strings.TrimSpace(token) == "" {
		return 0, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, fmt.Errorf("malformed page token")
	}
	value, ok := strings.CutPrefix(string(raw), "offset:")
	if !ok {
		return 0, fmt.Errorf("malformed page token")
	}
	offset, err := strconv.Atoi(value)
	if err != nil || offset < 0 {
		return 0, fmt.Errorf("malformed page token")
	}
	return offset, nil
}
