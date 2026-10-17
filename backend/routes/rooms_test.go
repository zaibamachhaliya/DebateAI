package routes

import (
	"strconv"
	"testing"
)

func TestGenerateRoomIDReturnsSixDigitCode(t *testing.T) {
	roomID, err := generateRoomID()
	if err != nil {
		t.Fatalf("generateRoomID() error = %v", err)
	}

	value, err := strconv.Atoi(roomID)
	if err != nil {
		t.Fatalf("generateRoomID() returned non-numeric code %q: %v", roomID, err)
	}
	if value < 100000 || value > 999999 {
		t.Fatalf("generateRoomID() = %q, want six-digit code", roomID)
	}
}
