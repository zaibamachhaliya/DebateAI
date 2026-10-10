package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JoinRequest struct {
	UserID    primitive.ObjectID `bson:"userId"`
	Status    string             `bson:"status"` // "pending", "approved", "rejected"
	CreatedAt time.Time          `bson:"createdAt"`
}

type Tournament struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty"`
	Title           string               `bson:"title"`
	Description     string               `bson:"description"`
	ModeratorName   string               `bson:"moderatorName"`
	ModeratorID     primitive.ObjectID   `bson:"moderatorId"`
	Category        string               `bson:"category"`
	Visibility      string               `bson:"visibility"`
	InviteCode      string               `bson:"inviteCode,omitempty"`
	MaxParticipants int                  `bson:"maxParticipants"`
	Participants    []primitive.ObjectID `bson:"participants"`
	JoinRequests    []JoinRequest        `bson:"joinRequests"`
	Status          string               `bson:"status"`
	ScheduledAt     time.Time            `bson:"scheduledAt"`
	CreatedAt       time.Time            `bson:"createdAt"`
}
