package websocket

import (
	"testing"

	gorilla "github.com/gorilla/websocket"
)

func TestBuildParticipantsMessageIncludesRecoverableRoomState(t *testing.T) {
	conn := &gorilla.Conn{}
	room := &Room{
		Clients: map[*gorilla.Conn]*Client{
			conn: {
				UserID:   "user-1",
				Username: "Alice",
				Email:    "alice@example.com",
				Role:     "for",
				Ready:    true,
			},
		},
	}

	message := buildParticipantsMessage(room)
	participants, ok := message["roomParticipants"].([]map[string]interface{})
	if !ok {
		t.Fatalf("roomParticipants has unexpected type %T", message["roomParticipants"])
	}
	if len(participants) != 1 {
		t.Fatalf("expected one participant, got %d", len(participants))
	}

	participant := participants[0]
	if ready, ok := participant["ready"].(bool); !ok || !ready {
		t.Fatalf("expected ready=true, got %#v", participant["ready"])
	}
	if role, ok := participant["role"].(string); !ok || role != "for" {
		t.Fatalf("expected role=for, got %#v", participant["role"])
	}
}

func TestCanStartDebateRequiresTwoReadyDebatersWithDifferentRoles(t *testing.T) {
	tests := []struct {
		name    string
		clients map[*gorilla.Conn]*Client
		want    bool
	}{
		{
			name: "different ready roles",
			clients: map[*gorilla.Conn]*Client{
				&gorilla.Conn{}: {UserID: "user-1", Role: "for", Ready: true},
				&gorilla.Conn{}: {UserID: "user-2", Role: "against", Ready: true},
			},
			want: true,
		},
		{
			name: "duplicate roles",
			clients: map[*gorilla.Conn]*Client{
				&gorilla.Conn{}: {UserID: "user-1", Role: "for", Ready: true},
				&gorilla.Conn{}: {UserID: "user-2", Role: "for", Ready: true},
			},
			want: false,
		},
		{
			name: "spectators do not count as debaters",
			clients: map[*gorilla.Conn]*Client{
				&gorilla.Conn{}: {UserID: "user-1", Role: "for", Ready: true},
				&gorilla.Conn{}: {UserID: "user-2", Role: "against", Ready: true},
				&gorilla.Conn{}: {UserID: "spectator", IsSpectator: true},
			},
			want: true,
		},
		{
			name: "unready debater cannot start",
			clients: map[*gorilla.Conn]*Client{
				&gorilla.Conn{}: {UserID: "user-1", Role: "for", Ready: true},
				&gorilla.Conn{}: {UserID: "user-2", Role: "against"},
			},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			room := &Room{Clients: test.clients}
			if got := canStartDebateLocked(room); got != test.want {
				t.Fatalf("canStartDebateLocked() = %t, want %t", got, test.want)
			}
		})
	}
}
