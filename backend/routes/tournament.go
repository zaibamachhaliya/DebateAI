package routes

import (
	"arguehub/controllers"

	"github.com/gin-gonic/gin"
)

func SetupTournamentRoutes(router *gin.RouterGroup) {
	// Public Tournament Join
	router.POST("/tournaments/:id/join", controllers.JoinPublicTournament)

	// Private Tournament Join (Request)
	router.POST("/tournaments/join", controllers.JoinPrivateTournament)

	// Moderator APIs
	router.GET("/tournaments/:id/requests", controllers.GetPendingRequests)
	router.POST("/tournaments/:id/requests/:userId/approve", controllers.ApproveJoinRequest)
	router.POST("/tournaments/:id/requests/:userId/reject", controllers.RejectJoinRequest)
}
