package domain

import (
	"testing"
	"time"
)

func TestUser(t *testing.T) {
	u := User{
		ID:        "1",
		Email:     "test@example.com",
		CreatedAt: time.Now(),
	}

	// Test fields are set correcty
	if u.ID != "1" {
		t.Errorf("Expected ID 1, got %s", u.ID)
	}
	if u.Email != "test@example.com" {
		t.Errorf("Expected Email test@example.com, got %s", u.Email)
	}
	if u.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}
