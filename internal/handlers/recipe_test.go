package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/culinaryshare/backend/internal/config"
	"github.com/culinaryshare/backend/internal/database"
	"github.com/culinaryshare/backend/internal/middleware"
	"github.com/culinaryshare/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper to create test JWT token
func createTestToken(userID, username, email, role string) string {
	if config.AppConfig == nil {
		config.AppConfig = &config.Config{
			JWTSecret:    "test-secret",
			JWTExpiresIn: 3600,
		}
	}

	claims := jwt.MapClaims{
		"userId":   userID,
		"username": username,
		"email":    email,
		"role":     role,
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(config.AppConfig.JWTSecret))
	return tokenString
}

// ==================== RECIPE CREATE TESTS ====================

func TestCreateRecipe_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create test category
	categoryID := primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	database.CategoriesCollection.InsertOne(ctx, models.Category{
		ID:        categoryID,
		Name:      "Italian",
		CreatedAt: time.Now(),
	})

	// Initialize config for JWT
	config.AppConfig = &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: 3600,
	}

	userID := primitive.NewObjectID()
	token := createTestToken(userID.Hex(), "testuser", "test@example.com", models.RoleUser)

	router := setupTestRouter()
	router.POST("/api/recipes", middleware.AuthRequired(), CreateRecipe)

	body := models.CreateRecipeInput{
		Title:        "Spaghetti Carbonara",
		Description:  "Classic Italian pasta dish",
		Instructions: "1. Cook pasta. 2. Fry bacon. 3. Mix eggs and cheese. 4. Combine.",
		CookingTime:  30,
		Servings:     4,
		CategoryID:   categoryID.Hex(),
		Ingredients: []models.Ingredient{
			{Name: "Spaghetti", Quantity: "400", Unit: "g"},
			{Name: "Eggs", Quantity: "4", Unit: "pieces"},
		},
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/recipes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["data"] == nil {
		t.Error("Expected data in response")
	}

	data := response["data"].(map[string]interface{})
	if data["title"] != "Spaghetti Carbonara" {
		t.Errorf("Expected title 'Spaghetti Carbonara', got %v", data["title"])
	}
}

func TestCreateRecipe_Unauthorized(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	config.AppConfig = &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: 3600,
	}

	router := setupTestRouter()
	router.POST("/api/recipes", middleware.AuthRequired(), CreateRecipe)

	body := models.CreateRecipeInput{
		Title:        "Test Recipe",
		Description:  "Test",
		Instructions: "Test",
		CookingTime:  30,
		Servings:     4,
		CategoryID:   primitive.NewObjectID().Hex(),
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/recipes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestCreateRecipe_InvalidCategory(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	config.AppConfig = &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: 3600,
	}

	userID := primitive.NewObjectID()
	token := createTestToken(userID.Hex(), "testuser", "test@example.com", models.RoleUser)

	router := setupTestRouter()
	router.POST("/api/recipes", middleware.AuthRequired(), CreateRecipe)

	body := models.CreateRecipeInput{
		Title:        "Test Recipe",
		Description:  "Test",
		Instructions: "Test",
		CookingTime:  30,
		Servings:     4,
		CategoryID:   primitive.NewObjectID().Hex(), // Non-existent category
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/recipes", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// ==================== REVIEW TESTS ====================

func TestAddReview_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	config.AppConfig = &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: 3600,
	}

	// Create test recipe
	recipeID := primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	database.RecipesCollection.InsertOne(ctx, models.Recipe{
		ID:            recipeID,
		Title:         "Test Recipe",
		Description:   "Test",
		Instructions:  "Test",
		Reviews:       []models.Review{},
		AverageRating: 0,
		ReviewCount:   0,
		CreatedAt:     time.Now(),
	})

	userID := primitive.NewObjectID()
	token := createTestToken(userID.Hex(), "testuser", "test@example.com", models.RoleUser)

	router := setupTestRouter()
	router.POST("/api/recipes/:id/reviews", middleware.AuthRequired(), AddReview)

	body := models.ReviewInput{
		Rating:  5,
		Comment: "Excellent recipe!",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/recipes/%s/reviews", recipeID.Hex()), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAddReview_InvalidRating(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	config.AppConfig = &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: 3600,
	}

	recipeID := primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	database.RecipesCollection.InsertOne(ctx, models.Recipe{
		ID:          recipeID,
		Title:       "Test Recipe",
		Reviews:     []models.Review{},
		ReviewCount: 0,
		CreatedAt:   time.Now(),
	})

	userID := primitive.NewObjectID()
	token := createTestToken(userID.Hex(), "testuser", "test@example.com", models.RoleUser)

	router := setupTestRouter()
	router.POST("/api/recipes/:id/reviews", middleware.AuthRequired(), AddReview)

	// Test rating > 5
	body := models.ReviewInput{
		Rating:  10,
		Comment: "Test",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/recipes/%s/reviews", recipeID.Hex()), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid rating, got %d", w.Code)
	}
}

func TestAddReview_DuplicateReview(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	config.AppConfig = &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: 3600,
	}

	userID := primitive.NewObjectID()
	recipeID := primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create recipe with existing review from user
	database.RecipesCollection.InsertOne(ctx, models.Recipe{
		ID:    recipeID,
		Title: "Test Recipe",
		Reviews: []models.Review{
			{
				ID:        primitive.NewObjectID(),
				UserID:    userID,
				Username:  "testuser",
				Rating:    4,
				Comment:   "Good",
				CreatedAt: time.Now(),
			},
		},
		AverageRating: 4,
		ReviewCount:   1,
		CreatedAt:     time.Now(),
	})

	token := createTestToken(userID.Hex(), "testuser", "test@example.com", models.RoleUser)

	router := setupTestRouter()
	router.POST("/api/recipes/:id/reviews", middleware.AuthRequired(), AddReview)

	body := models.ReviewInput{
		Rating:  5,
		Comment: "Another review",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/recipes/%s/reviews", recipeID.Hex()), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 for duplicate review, got %d", w.Code)
	}
}

func TestAddReview_AtomicRatingUpdate(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	config.AppConfig = &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: 3600,
	}

	recipeID := primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create recipe with one existing review
	existingUserID := primitive.NewObjectID()
	database.RecipesCollection.InsertOne(ctx, models.Recipe{
		ID:    recipeID,
		Title: "Test Recipe",
		Reviews: []models.Review{
			{
				ID:        primitive.NewObjectID(),
				UserID:    existingUserID,
				Username:  "existinguser",
				Rating:    4,
				Comment:   "Good",
				CreatedAt: time.Now(),
			},
		},
		AverageRating: 4.0,
		ReviewCount:   1,
		CreatedAt:     time.Now(),
	})

	// Add a new review
	newUserID := primitive.NewObjectID()
	token := createTestToken(newUserID.Hex(), "newuser", "new@example.com", models.RoleUser)

	router := setupTestRouter()
	router.POST("/api/recipes/:id/reviews", middleware.AuthRequired(), AddReview)

	body := models.ReviewInput{
		Rating:  2,
		Comment: "Not great",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/recipes/%s/reviews", recipeID.Hex()), bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the average rating was updated atomically
	// (4 + 2) / 2 = 3.0
	var updatedRecipe models.Recipe
	database.RecipesCollection.FindOne(ctx, primitive.M{"_id": recipeID}).Decode(&updatedRecipe)

	if updatedRecipe.ReviewCount != 2 {
		t.Errorf("Expected reviewCount 2, got %d", updatedRecipe.ReviewCount)
	}

	expectedAvg := 3.0
	if updatedRecipe.AverageRating != expectedAvg {
		t.Errorf("Expected averageRating %.1f, got %.1f", expectedAvg, updatedRecipe.AverageRating)
	}
}

func TestDeleteReview_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	config.AppConfig = &config.Config{
		JWTSecret:    "test-secret",
		JWTExpiresIn: 3600,
	}

	userID := primitive.NewObjectID()
	recipeID := primitive.NewObjectID()
	reviewID := primitive.NewObjectID()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create recipe with a review
	database.RecipesCollection.InsertOne(ctx, models.Recipe{
		ID:    recipeID,
		Title: "Test Recipe",
		Reviews: []models.Review{
			{
				ID:        reviewID,
				UserID:    userID,
				Username:  "testuser",
				Rating:    5,
				Comment:   "Great",
				CreatedAt: time.Now(),
			},
		},
		AverageRating: 5.0,
		ReviewCount:   1,
		CreatedAt:     time.Now(),
	})

	token := createTestToken(userID.Hex(), "testuser", "test@example.com", models.RoleUser)

	router := setupTestRouter()
	router.DELETE("/api/recipes/:id/reviews/:reviewId", middleware.AuthRequired(), DeleteReview)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/recipes/%s/reviews/%s", recipeID.Hex(), reviewID.Hex()), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify review was deleted and rating reset
	var updatedRecipe models.Recipe
	database.RecipesCollection.FindOne(ctx, primitive.M{"_id": recipeID}).Decode(&updatedRecipe)

	if updatedRecipe.ReviewCount != 0 {
		t.Errorf("Expected reviewCount 0, got %d", updatedRecipe.ReviewCount)
	}

	if updatedRecipe.AverageRating != 0 {
		t.Errorf("Expected averageRating 0, got %.1f", updatedRecipe.AverageRating)
	}
}
