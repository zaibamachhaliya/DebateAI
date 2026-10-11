package controllers

import (
	"arguehub/db"
	"arguehub/models"
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ============================================
// HELPER FUNCTION
// ============================================

func getUserIDFromContext(c *gin.Context) (primitive.ObjectID, error) {
	userIDRaw, exists := c.Get("userID")
	if !exists {
		return primitive.NilObjectID, mongo.ErrNoDocuments
	}
	userObjID, ok := userIDRaw.(primitive.ObjectID)
	if !ok {
		return primitive.NilObjectID, mongo.ErrNoDocuments
	}
	return userObjID, nil
}

// ============================================
// PUBLIC TOURNAMENT JOIN
// ============================================

func JoinPublicTournament(c *gin.Context) {
	tournamentID := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(tournamentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}

	userObjID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var tournament models.Tournament
	err = db.TournamentCollection.FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&tournament)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	if tournament.Visibility != "public" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This is a private tournament. Please use invite code."})
		return
	}

	if len(tournament.Participants) >= tournament.MaxParticipants {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament is full"})
		return
	}

	for _, p := range tournament.Participants {
		if p.UserID.Hex() == userObjID.Hex() { // ← YEH CHANGE KARO
			c.JSON(http.StatusBadRequest, gin.H{"error": "You are already in this tournament"})
			return
		}
	}

	newParticipant := models.TournamentParticipant{
		UserID: userObjID,
		Name:   "", // Ya user ka naam
		Stance: "",
	}

	_, err = db.TournamentCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$push": bson.M{"participants": newParticipant}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join tournament"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Joined successfully",
		"tournamentId": tournamentID,
		"userId":       userObjID.Hex(),
	})
}

// ============================================
// PRIVATE TOURNAMENT JOIN (REQUEST)
// ============================================

func JoinPrivateTournament(c *gin.Context) {
	var req struct {
		InviteCode string `json:"inviteCode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invite code is required"})
		return
	}

	userObjID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var tournament models.Tournament
	err = db.TournamentCollection.FindOne(
		context.Background(),
		bson.M{"inviteCode": req.InviteCode, "visibility": "private"},
	).Decode(&tournament)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid invite code"})
		return
	}

	if len(tournament.Participants) >= tournament.MaxParticipants {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament is full"})
		return
	}

	for _, p := range tournament.Participants {
		if p.UserID.Hex() == userObjID.Hex() { // ← YEH CHANGE KARO
			c.JSON(http.StatusBadRequest, gin.H{"error": "You are already in this tournament"})
			return
		}
	}

	// Existing request check
	for _, r := range tournament.JoinRequests {
		if r.UserID.Hex() == userObjID.Hex() {
			if r.Status == "pending" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "You have already requested to join"})
				return
			}
			if r.Status == "approved" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Your request was already approved"})
				return
			}
			// Rejected - purana request hatao, naya bhejne do
			_, err = db.TournamentCollection.UpdateOne(
				context.Background(),
				bson.M{"_id": tournament.ID},
				bson.M{"$pull": bson.M{"joinRequests": bson.M{"userId": userObjID}}},
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
				return
			}
			break
		}
	}

	newRequest := models.JoinRequest{
		UserID:    userObjID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	_, err = db.TournamentCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": tournament.ID},
		bson.M{"$push": bson.M{"joinRequests": newRequest}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send join request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Request sent to moderator. Please wait for approval.",
		"tournamentId": tournament.ID.Hex(),
	})
}

// ============================================
// MODERATOR: GET PENDING REQUESTS
// ============================================

func GetPendingRequests(c *gin.Context) {
	tournamentID := c.Param("id")

	objID, err := primitive.ObjectIDFromHex(tournamentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}

	userObjID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var tournament models.Tournament
	err = db.TournamentCollection.FindOne(
		context.Background(),
		bson.M{"_id": objID, "moderatorId": userObjID},
	).Decode(&tournament)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{"error": "You are not the moderator of this tournament"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tournament"})
		return
	}

	var pendingRequests []models.JoinRequest
	for _, r := range tournament.JoinRequests {
		if r.Status == "pending" {
			pendingRequests = append(pendingRequests, r)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"pendingRequests": pendingRequests,
		"count":           len(pendingRequests),
		"tournamentId":    tournamentID,
		"participants":    len(tournament.Participants),
		"maxParticipants": tournament.MaxParticipants,
	})
}

// ============================================
// CHOOSE STANCE
// ============================================

func ChooseStance(c *gin.Context) {
	tournamentID := c.Param("id")

	var req struct {
		Stance string `json:"stance" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stance is required"})
		return
	}

	// Validate stance
	if req.Stance != "for" && req.Stance != "against" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stance must be 'for' or 'against'"})
		return
	}

	userObjID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(tournamentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}

	// Check user is participant
	var tournament models.Tournament
	err = db.TournamentCollection.FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&tournament)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	// Check user in participants
	isParticipant := false
	for _, p := range tournament.Participants {
		if p.UserID.Hex() == userObjID.Hex() {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not a participant of this tournament"})
		return
	}

	// Update stance
	_, err = db.TournamentCollection.UpdateOne(
		context.Background(),
		bson.M{
			"_id":                 objID,
			"participants.userId": userObjID,
		},
		bson.M{"$set": bson.M{"participants.$.stance": req.Stance}},
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stance updated successfully",
		"stance":  req.Stance,
	})
}

// ============================================
// MODERATOR: MANAGE REQUEST (APPROVE/REJECT)
// ============================================

func ManageJoinRequest(c *gin.Context) {
	tournamentID := c.Param("id")
	requestUserID := c.Param("userId")

	// Body se action lo (approve ya reject)
	var req struct {
		Action string `json:"action" binding:"required"` // "approve" ya "reject"
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action is required (approve/reject)"})
		return
	}

	// Validate action
	if req.Action != "approve" && req.Action != "reject" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Action must be 'approve' or 'reject'"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(tournamentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tournament ID"})
		return
	}

	requestUserObjID, err := primitive.ObjectIDFromHex(requestUserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	userObjID, err := getUserIDFromContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var tournament models.Tournament
	err = db.TournamentCollection.FindOne(
		context.Background(),
		bson.M{"_id": objID},
	).Decode(&tournament)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tournament not found"})
		return
	}

	// Check moderator
	if tournament.ModeratorID.Hex() != userObjID.Hex() {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only moderator can manage requests"})
		return
	}

	// Request exist check
	requestFound := false
	for _, r := range tournament.JoinRequests {
		if r.UserID.Hex() == requestUserObjID.Hex() && r.Status == "pending" {
			requestFound = true
			break
		}
	}

	if !requestFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending request not found"})
		return
	}

	// ========== APPROVE ==========
	if req.Action == "approve" {
		// Full check (mentor ne bola tha)
		if len(tournament.Participants) >= tournament.MaxParticipants {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tournament is full. Cannot approve more requests."})
			return
		}

		newParticipant := models.TournamentParticipant{
			UserID: requestUserObjID,
			Name:   "",
			Stance: "",
		}

		_, err = db.TournamentCollection.UpdateOne(
			context.Background(),
			bson.M{"_id": objID},
			bson.M{
				"$push": bson.M{"participants": newParticipant},
				"$set":  bson.M{"joinRequests.$[elem].status": "approved"},
			},
			options.Update().SetArrayFilters(options.ArrayFilters{
				Filters: []interface{}{
					bson.M{"elem.userId": requestUserObjID, "elem.status": "pending"},
				},
			}),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve request"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Request approved successfully",
			"userId":  requestUserID,
			"action":  "approve",
		})
		return
	}

	// ========== REJECT ==========
	if req.Action == "reject" {
		result, err := db.TournamentCollection.UpdateOne(
			context.Background(),
			bson.M{
				"_id": objID,
				"joinRequests": bson.M{
					"$elemMatch": bson.M{"userId": requestUserObjID, "status": "pending"},
				},
			},
			bson.M{"$set": bson.M{"joinRequests.$[elem].status": "rejected"}},
			options.Update().SetArrayFilters(options.ArrayFilters{
				Filters: []interface{}{
					bson.M{"elem.userId": requestUserObjID, "elem.status": "pending"},
				},
			}),
		)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject request"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Pending request not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Request rejected successfully",
			"userId":  requestUserID,
			"action":  "reject",
		})
		return
	}
}
