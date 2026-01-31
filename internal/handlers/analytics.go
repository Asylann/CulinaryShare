package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/culinaryshare/backend/internal/database"
	"github.com/culinaryshare/backend/internal/models"
	"github.com/culinaryshare/backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// GetTopRatedRecipes returns top-rated recipes using aggregation pipeline
func GetTopRatedRecipes(c *gin.Context) {
	// Parse limit from query (default 10)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Multi-stage Aggregation Pipeline:
	// 1. $match - Filter recipes that have at least one review
	// 2. $lookup - Join with categories collection to get category name
	// 3. $lookup - Join with users collection to get author name
	// 4. $unwind - Flatten the category array from $lookup
	// 5. $unwind - Flatten the user array from $lookup
	// 6. $project - Select and rename fields
	// 7. $sort - Sort by averageRating descending
	// 8. $limit - Limit results

	pipeline := []bson.M{
		// Stage 1: Match recipes with reviews
		{
			"$match": bson.M{
				"reviewCount": bson.M{"$gt": 0},
			},
		},
		// Stage 2: Lookup categories
		{
			"$lookup": bson.M{
				"from":         "categories",
				"localField":   "categoryId",
				"foreignField": "_id",
				"as":           "category",
			},
		},
		// Stage 3: Lookup users (recipe authors)
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "userId",
				"foreignField": "_id",
				"as":           "author",
			},
		},
		// Stage 4: Unwind category array
		{
			"$unwind": bson.M{
				"path":                       "$category",
				"preserveNullAndEmptyArrays": true,
			},
		},
		// Stage 5: Unwind author array
		{
			"$unwind": bson.M{
				"path":                       "$author",
				"preserveNullAndEmptyArrays": true,
			},
		},
		// Stage 6: Project required fields
		{
			"$project": bson.M{
				"_id":           1,
				"title":         1,
				"description":   1,
				"averageRating": 1,
				"reviewCount":   1,
				"categoryName":  "$category.name",
				"username":      "$author.username",
			},
		},
		// Stage 7: Sort by average rating (descending)
		{
			"$sort": bson.M{
				"averageRating": -1,
				"reviewCount":   -1,
			},
		},
		// Stage 8: Limit results
		{
			"$limit": limit,
		},
	}

	cursor, err := database.RecipesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		utils.InternalServerError(c, "Failed to aggregate recipes", err)
		return
	}
	defer cursor.Close(ctx)

	var topRatedRecipes []models.TopRatedRecipe
	if err := cursor.All(ctx, &topRatedRecipes); err != nil {
		utils.InternalServerError(c, "Failed to decode aggregation results", err)
		return
	}

	if topRatedRecipes == nil {
		topRatedRecipes = []models.TopRatedRecipe{}
	}

	utils.OK(c, "Top rated recipes retrieved successfully", topRatedRecipes)
}
