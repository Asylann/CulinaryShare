package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Ingredient represents an ingredient in a recipe (EMBEDDED document)
type Ingredient struct {
	Name     string `bson:"name" json:"name"`
	Quantity string `bson:"quantity" json:"quantity"`
	Unit     string `bson:"unit" json:"unit"`
}

// Review represents a review on a recipe (EMBEDDED document)
type Review struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"userId" json:"userId"`
	Username  string             `bson:"username" json:"username"`
	Rating    int                `bson:"rating" json:"rating"` // 1-5
	Comment   string             `bson:"comment" json:"comment"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
}

// Recipe represents a recipe in the system
type Recipe struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title         string             `bson:"title" json:"title"`
	Description   string             `bson:"description" json:"description"`
	Ingredients   []Ingredient       `bson:"ingredients" json:"ingredients"` // EMBEDDED
	Instructions  string             `bson:"instructions" json:"instructions"`
	CookingTime   int                `bson:"cookingTime" json:"cookingTime"` // in minutes
	Servings      int                `bson:"servings" json:"servings"`
	CategoryID    primitive.ObjectID `bson:"categoryId" json:"categoryId"` // REFERENCED
	UserID        primitive.ObjectID `bson:"userId" json:"userId"`         // REFERENCED
	Reviews       []Review           `bson:"reviews" json:"reviews"`       // EMBEDDED
	AverageRating float64            `bson:"averageRating" json:"averageRating"`
	ReviewCount   int                `bson:"reviewCount" json:"reviewCount"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// CreateRecipeInput represents the input for creating a recipe
type CreateRecipeInput struct {
	Title        string       `json:"title" binding:"required"`
	Description  string       `json:"description" binding:"required"`
	Ingredients  []Ingredient `json:"ingredients"`
	Instructions string       `json:"instructions" binding:"required"`
	CookingTime  int          `json:"cookingTime" binding:"required"`
	Servings     int          `json:"servings" binding:"required"`
	CategoryID   string       `json:"categoryId" binding:"required"`
}

// UpdateRecipeInput represents the input for updating a recipe
type UpdateRecipeInput struct {
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	Ingredients  []Ingredient `json:"ingredients"`
	Instructions string       `json:"instructions"`
	CookingTime  int          `json:"cookingTime"`
	Servings     int          `json:"servings"`
	CategoryID   string       `json:"categoryId"`
}

// IngredientInput represents the input for adding/removing ingredients
type IngredientInput struct {
	Action      string       `json:"action" binding:"required"` // "add" or "remove"
	Ingredients []Ingredient `json:"ingredients" binding:"required"`
}

// ReviewInput represents the input for creating a review
type ReviewInput struct {
	Rating  int    `json:"rating" binding:"required"`
	Comment string `json:"comment"`
}

// RecipeWithCategory represents a recipe with populated category info
type RecipeWithCategory struct {
	Recipe       `bson:",inline"`
	CategoryName string `bson:"categoryName" json:"categoryName"`
}

// TopRatedRecipe represents the aggregation result for top-rated recipes
type TopRatedRecipe struct {
	ID            primitive.ObjectID `bson:"_id" json:"id"`
	Title         string             `bson:"title" json:"title"`
	Description   string             `bson:"description" json:"description"`
	AverageRating float64            `bson:"averageRating" json:"averageRating"`
	ReviewCount   int                `bson:"reviewCount" json:"reviewCount"`
	CategoryName  string             `bson:"categoryName" json:"categoryName"`
	Username      string             `bson:"username" json:"username"`
}
