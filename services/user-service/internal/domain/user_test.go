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

	if u.ID != "1" {
		t.Errorf("Expected ID 1, got %s", u.ID)
	}
}
