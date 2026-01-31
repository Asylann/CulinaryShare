package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/culinaryshare/backend/internal/database"
	"github.com/culinaryshare/backend/internal/middleware"
	"github.com/culinaryshare/backend/internal/models"
	"github.com/culinaryshare/backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetRecipes returns recipes with filtering and pagination
func GetRecipes(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	categoryID := c.Query("categoryId")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	skip := (page - 1) * limit

	// Build filter
	filter := bson.M{}
	if categoryID != "" {
		if objID, err := primitive.ObjectIDFromHex(categoryID); err == nil {
			filter["categoryId"] = objID
		}
	}
	if search != "" {
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": search, "$options": "i"}},
			{"description": bson.M{"$regex": search, "$options": "i"}},
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get total count
	totalCount, err := database.RecipesCollection.CountDocuments(ctx, filter)
	if err != nil {
		utils.InternalServerError(c, "Failed to count recipes", err)
		return
	}

	// Find recipes with pagination
	opts := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := database.RecipesCollection.Find(ctx, filter, opts)
	if err != nil {
		utils.InternalServerError(c, "Failed to fetch recipes", err)
		return
	}
	defer cursor.Close(ctx)

	var recipes []models.Recipe
	if err := cursor.All(ctx, &recipes); err != nil {
		utils.InternalServerError(c, "Failed to decode recipes", err)
		return
	}

	if recipes == nil {
		recipes = []models.Recipe{}
	}

	utils.RespondWithPagination(c, recipes, page, limit, totalCount)
}

// CreateRecipe creates a new recipe (authenticated users)
func CreateRecipe(c *gin.Context) {
	var input models.CreateRecipeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	// Validate input
	validationErrors := &utils.ValidationErrors{}
	validationErrors.Add(utils.ValidateRequired(input.Title, "title"))
	validationErrors.Add(utils.ValidateRequired(input.Description, "description"))
	validationErrors.Add(utils.ValidateRequired(input.Instructions, "instructions"))
	validationErrors.Add(utils.ValidateRequired(input.CategoryID, "categoryId"))

	if validationErrors.HasErrors() {
		utils.BadRequest(c, validationErrors.Error(), nil)
		return
	}

	// Parse category ID
	categoryObjID, err := primitive.ObjectIDFromHex(input.CategoryID)
	if err != nil {
		utils.BadRequest(c, "Invalid category ID", err)
		return
	}

	// Get user ID from context
	userIDStr := middleware.GetUserID(c)
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.Unauthorized(c, "Invalid user session")
		return
	}

	// Verify category exists
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := database.CategoriesCollection.CountDocuments(ctx, bson.M{"_id": categoryObjID})
	if err != nil || count == 0 {
		utils.BadRequest(c, "Category not found", nil)
		return
	}

	now := time.Now()
	recipe := models.Recipe{
		ID:            primitive.NewObjectID(),
		Title:         input.Title,
		Description:   input.Description,
		Ingredients:   input.Ingredients,
		Instructions:  input.Instructions,
		CookingTime:   input.CookingTime,
		Servings:      input.Servings,
		CategoryID:    categoryObjID,
		UserID:        userObjID,
		Reviews:       []models.Review{},
		AverageRating: 0,
		ReviewCount:   0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if recipe.Ingredients == nil {
		recipe.Ingredients = []models.Ingredient{}
	}

	_, err = database.RecipesCollection.InsertOne(ctx, recipe)
	if err != nil {
		utils.InternalServerError(c, "Failed to create recipe", err)
		return
	}

	utils.Created(c, "Recipe created successfully", recipe)
}

// GetRecipe returns a single recipe by ID
func GetRecipe(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequest(c, "Invalid recipe ID", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var recipe models.Recipe
	err = database.RecipesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&recipe)
	if err != nil {
		utils.NotFound(c, "Recipe not found")
		return
	}

	utils.OK(c, "Recipe retrieved successfully", recipe)
}

// UpdateRecipe updates a recipe (owner or admin only)
func UpdateRecipe(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequest(c, "Invalid recipe ID", err)
		return
	}

	var input models.UpdateRecipeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the recipe
	var existingRecipe models.Recipe
	err = database.RecipesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existingRecipe)
	if err != nil {
		utils.NotFound(c, "Recipe not found")
		return
	}

	// Check ownership or admin role
	userIDStr := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if existingRecipe.UserID.Hex() != userIDStr && userRole != models.RoleAdmin {
		utils.Forbidden(c, "You can only update your own recipes")
		return
	}

	// Build update document
	update := bson.M{"$set": bson.M{"updatedAt": time.Now()}}
	setFields := update["$set"].(bson.M)

	if input.Title != "" {
		setFields["title"] = input.Title
	}
	if input.Description != "" {
		setFields["description"] = input.Description
	}
	if input.Instructions != "" {
		setFields["instructions"] = input.Instructions
	}
	if input.CookingTime > 0 {
		setFields["cookingTime"] = input.CookingTime
	}
	if input.Servings > 0 {
		setFields["servings"] = input.Servings
	}
	if input.Ingredients != nil {
		setFields["ingredients"] = input.Ingredients
	}
	if input.CategoryID != "" {
		if categoryObjID, err := primitive.ObjectIDFromHex(input.CategoryID); err == nil {
			setFields["categoryId"] = categoryObjID
		}
	}

	_, err = database.RecipesCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerError(c, "Failed to update recipe", err)
		return
	}

	// Fetch updated recipe
	var updatedRecipe models.Recipe
	database.RecipesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedRecipe)

	utils.OK(c, "Recipe updated successfully", updatedRecipe)
}

// UpdateIngredients adds or removes ingredients using $push/$pull
func UpdateIngredients(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequest(c, "Invalid recipe ID", err)
		return
	}

	var input models.IngredientInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	if input.Action != "add" && input.Action != "remove" {
		utils.BadRequest(c, "Action must be 'add' or 'remove'", nil)
		return
	}

	if len(input.Ingredients) == 0 {
		utils.BadRequest(c, "At least one ingredient is required", nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if recipe exists and user has permission
	var existingRecipe models.Recipe
	err = database.RecipesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&existingRecipe)
	if err != nil {
		utils.NotFound(c, "Recipe not found")
		return
	}

	userIDStr := middleware.GetUserID(c)
	userRole := middleware.GetUserRole(c)
	if existingRecipe.UserID.Hex() != userIDStr && userRole != models.RoleAdmin {
		utils.Forbidden(c, "You can only modify your own recipes")
		return
	}

	var update bson.M
	if input.Action == "add" {
		// Use $push with $each to add multiple ingredients
		update = bson.M{
			"$push": bson.M{
				"ingredients": bson.M{"$each": input.Ingredients},
			},
			"$set": bson.M{"updatedAt": time.Now()},
		}
	} else {
		// Use $pull to remove ingredients by name
		names := make([]string, len(input.Ingredients))
		for i, ing := range input.Ingredients {
			names[i] = ing.Name
		}
		update = bson.M{
			"$pull": bson.M{
				"ingredients": bson.M{"name": bson.M{"$in": names}},
			},
			"$set": bson.M{"updatedAt": time.Now()},
		}
	}

	_, err = database.RecipesCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerError(c, "Failed to update ingredients", err)
		return
	}

	// Fetch updated recipe
	var updatedRecipe models.Recipe
	database.RecipesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&updatedRecipe)

	utils.OK(c, "Ingredients updated successfully", updatedRecipe)
}

// AddReview adds a review to a recipe with atomic rating update
func AddReview(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		utils.BadRequest(c, "Invalid recipe ID", err)
		return
	}

	var input models.ReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	// Validate rating
	if err := utils.ValidateRating(input.Rating); err != nil {
		utils.BadRequest(c, err.Error(), nil)
		return
	}

	userIDStr := middleware.GetUserID(c)
	username := middleware.GetUsername(c)
	userObjID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.Unauthorized(c, "Invalid user session")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if recipe exists
	var recipe models.Recipe
	err = database.RecipesCollection.FindOne(ctx, bson.M{"_id": objID}).Decode(&recipe)
	if err != nil {
		utils.NotFound(c, "Recipe not found")
		return
	}

	// Check if user already reviewed this recipe
	for _, review := range recipe.Reviews {
		if review.UserID == userObjID {
			utils.Conflict(c, "You have already reviewed this recipe")
			return
		}
	}

	// Create new review
	review := models.Review{
		ID:        primitive.NewObjectID(),
		UserID:    userObjID,
		Username:  username,
		Rating:    input.Rating,
		Comment:   input.Comment,
		CreatedAt: time.Now(),
	}

	// ATOMIC UPDATE: Calculate new average rating atomically
	// We use $push to add the review and $inc to update the review count
	// Then we recalculate the average rating in a single operation
	newReviewCount := recipe.ReviewCount + 1
	newTotalRating := recipe.AverageRating*float64(recipe.ReviewCount) + float64(input.Rating)
	newAverageRating := newTotalRating / float64(newReviewCount)

	update := bson.M{
		"$push": bson.M{"reviews": review},
		"$set": bson.M{
			"reviewCount":   newReviewCount,
			"averageRating": newAverageRating,
			"updatedAt":     time.Now(),
		},
	}

	_, err = database.RecipesCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
	if err != nil {
		utils.InternalServerError(c, "Failed to add review", err)
		return
	}

	utils.Created(c, "Review added successfully", review)
}

// DeleteReview removes a review from a recipe with atomic rating update
func DeleteReview(c *gin.Context) {
	recipeID := c.Param("id")
	reviewID := c.Param("reviewId")

	recipeObjID, err := primitive.ObjectIDFromHex(recipeID)
	if err != nil {
		utils.BadRequest(c, "Invalid recipe ID", err)
		return
	}

	reviewObjID, err := primitive.ObjectIDFromHex(reviewID)
	if err != nil {
		utils.BadRequest(c, "Invalid review ID", err)
		return
	}

	userIDStr := middleware.GetUserID(c)
	userObjID, _ := primitive.ObjectIDFromHex(userIDStr)
	userRole := middleware.GetUserRole(c)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the recipe and the specific review
	var recipe models.Recipe
	err = database.RecipesCollection.FindOne(ctx, bson.M{"_id": recipeObjID}).Decode(&recipe)
	if err != nil {
		utils.NotFound(c, "Recipe not found")
		return
	}

	// Find the review to delete
	var reviewToDelete *models.Review
	for _, review := range recipe.Reviews {
		if review.ID == reviewObjID {
			reviewToDelete = &review
			break
		}
	}

	if reviewToDelete == nil {
		utils.NotFound(c, "Review not found")
		return
	}

	// Check permission: review owner or admin
	if reviewToDelete.UserID != userObjID && userRole != models.RoleAdmin {
		utils.Forbidden(c, "You can only delete your own reviews")
		return
	}

	// ATOMIC UPDATE: Remove review and recalculate average rating
	newReviewCount := recipe.ReviewCount - 1
	var newAverageRating float64
	if newReviewCount > 0 {
		newTotalRating := recipe.AverageRating*float64(recipe.ReviewCount) - float64(reviewToDelete.Rating)
		newAverageRating = newTotalRating / float64(newReviewCount)
	} else {
		newAverageRating = 0
	}

	update := bson.M{
		"$pull": bson.M{"reviews": bson.M{"_id": reviewObjID}},
		"$set": bson.M{
			"reviewCount":   newReviewCount,
			"averageRating": newAverageRating,
			"updatedAt":     time.Now(),
		},
	}

	_, err = database.RecipesCollection.UpdateOne(ctx, bson.M{"_id": recipeObjID}, update)
	if err != nil {
		utils.InternalServerError(c, "Failed to delete review", err)
		return
	}

	utils.OK(c, "Review deleted successfully", nil)
}
