package util

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ParseTimestamp parses a timestamp string in RFC3339 format to a protobuf timestamp.
func ParseTimestamp(ts string) (*timestamppb.Timestamp, error) {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timestamp: %w", err)
	}
	return timestamppb.New(t), nil
}

// FormatTime formats a protobuf timestamp to RFC3339 string.
func FormatTime(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339)
}

func DerefString(s *string) string {
	if s == nil {
		return "" // Return zero value if pointer is nil
	}
	return *s // Return the underlying string value
}
