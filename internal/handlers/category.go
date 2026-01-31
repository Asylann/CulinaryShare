package handlers

import (
	"context"
	"time"

	"github.com/culinaryshare/backend/internal/database"
	"github.com/culinaryshare/backend/internal/models"
	"github.com/culinaryshare/backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetCategories returns all categories
func GetCategories(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := database.CategoriesCollection.Find(ctx, bson.M{})
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch categories", err)
		return
	}
	defer cursor.Close(ctx)

	var categories []models.Category
	if err := cursor.All(ctx, &categories); err != nil {
		utils.InternalServerError(c, "Failed to decode categories", err)
		return
	}

	if categories == nil {
		categories = []models.Category{}
	}

	utils.OK(c, "Categories retrieved successfully", categories)
}

// CreateCategory creates a new category (ADMIN only)
func CreateCategory(c *gin.Context) {
	var input models.CreateCategoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	// Validate input
	if err := utils.ValidateRequired(input.Name, "name"); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	now := time.Now()
	category := models.Category{
		ID:          primitive.NewObjectID(),
		Name:        input.Name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := database.CategoriesCollection.InsertOne(ctx, category)
	if err != nil {
		utils.InternalServerError(c, "Failed to create category", err)
		return
	}

	utils.Created(c, "Category created successfully", category)
}
