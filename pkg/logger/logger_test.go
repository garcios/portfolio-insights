package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	l := New()
	if l == nil {
		t.Error("Expected logger to be non-nil")
	}
}
