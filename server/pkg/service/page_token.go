package service

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// pageTokenCodec converts numeric list offsets into v1 page tokens. The
// token is a reversible base64-wrapped offset — opaque to casual eyeballs,
// but not a secret and trivially decodable; it carries no authority, so a
// forged or stale token at worst yields an empty or repeated page. The
// offset addresses the live ordering, not a snapshot: rows that appear or
// leave the filtered set mid-walk shift positions, so pages may skip or
// repeat under mutation. Changing the wire format invalidates outstanding
// tokens, which is acceptable — clients simply restart from the first
// page.
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
// A malformed token is invalid caller input and carries the
// ErrInvalidArgument sentinel so list handlers translate it through the
// shared taxonomy.
func (pageTokenCodecType) Decode(token string) (int, error) {
	if strings.TrimSpace(token) == "" {
		return 0, nil
	}
	malformed := wrapf(ErrInvalidArgument, "malformed page token")
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0, malformed
	}
	value, ok := strings.CutPrefix(string(raw), "offset:")
	if !ok {
		return 0, malformed
	}
	offset, err := strconv.Atoi(value)
	if err != nil || offset < 0 {
		return 0, malformed
	}
	return offset, nil
}
