package utils_test

import (
	"testing"
	"time"

	"github.com/your-team/koala-exam-backend/pkg/utils"
)

func TestUUID(t *testing.T) {
	id := utils.UUID()
	if len(id) < 30 { t.Fatal("uuid too short") }
}

func TestRandString(t *testing.T) {
	s := utils.RandString(8)
	// RandString returns hex (2 chars per byte), so length is 2*n
	if len(s) != 16 { t.Fatalf("expected 16, got %d", len(s)) }
}

func TestFormatTime(t *testing.T) {
	now := time.Now()
	s := utils.FormatTime(now)
	if len(s) != 19 { t.Fatalf("expected 19 chars, got %d", len(s)) }
}
