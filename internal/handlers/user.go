package handlers

import (
	"context"
	"time"

	"github.com/culinaryshare/backend/internal/database"
	"github.com/culinaryshare/backend/internal/middleware"
	"github.com/culinaryshare/backend/internal/models"
	"github.com/culinaryshare/backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetMe returns the current user's profile
func GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == "" {
		utils.Unauthorized(c, "User not authenticated")
		return
	}

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		utils.BadRequest(c, "Invalid user ID", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err = database.UsersCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		utils.NotFound(c, "User not found")
		return
	}

	// Return user without password (using ToResponse)
	utils.OK(c, "User profile retrieved successfully", user.ToResponse())
}
