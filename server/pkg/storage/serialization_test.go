package storage

import (
	"errors"
	"fmt"
	"testing"
)

type sqlStateError struct {
	state string
}

func (e *sqlStateError) SQLState() string { return e.state }

func (e *sqlStateError) Error() string {
	return fmt.Sprintf("sqlstate error: %s", e.state)
}

func TestIsSerializationFailure(t *testing.T) {
	cases := []struct {
		err  error
		want bool
	}{
		{&sqlStateError{state: "40001"}, true},
		{&sqlStateError{state: "40000"}, false},
		{&sqlStateError{state: "23505"}, false},
		{errors.New("plain error"), false},
		{fmt.Errorf("wrapped: %w", &sqlStateError{state: "40001"}), true},
		{nil, false},
	}
	for _, tc := range cases {
		if got := isSerializationFailure(tc.err); got != tc.want {
			t.Fatalf("isSerializationFailure(%v) = %v, want %v", tc.err, got, tc.want)
		}
	}
}
