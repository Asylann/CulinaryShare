package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/culinaryshare/backend/internal/database"
	"github.com/culinaryshare/backend/internal/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// Test configuration
var testDB *mongo.Database

func setupTestDB(t *testing.T) {
	// Use MongoDB memory server or test database
	// For unit tests, we'll use a test database
	uri := "mongodb://localhost:27017"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Skip("MongoDB not available, skipping integration tests")
		return
	}

	testDB = client.Database("culinaryshare_test")
	database.UsersCollection = testDB.Collection("users")
	database.RecipesCollection = testDB.Collection("recipes")
	database.CategoriesCollection = testDB.Collection("categories")
}

func teardownTestDB(t *testing.T) {
	if testDB != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		testDB.Drop(ctx)
	}
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// ==================== AUTH TESTS ====================

func TestRegister_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupTestRouter()
	router.POST("/api/auth/register", Register)

	body := models.RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
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
	if data["token"] == nil {
		t.Error("Expected token in response")
	}
	if data["user"] == nil {
		t.Error("Expected user in response")
	}

	// Verify password is not returned
	user := data["user"].(map[string]interface{})
	if user["password"] != nil {
		t.Error("Password should not be returned in response")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create unique index on email
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	database.UsersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	router := setupTestRouter()
	router.POST("/api/auth/register", Register)

	body := models.RegisterInput{
		Username: "testuser1",
		Email:    "duplicate@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	// First registration
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("First registration failed: %s", w.Body.String())
	}

	// Second registration with same email
	body.Username = "testuser2"
	jsonBody, _ = json.Marshal(body)
	req, _ = http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409 for duplicate email, got %d", w.Code)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupTestRouter()
	router.POST("/api/auth/register", Register)

	body := models.RegisterInput{
		Username: "testuser",
		Email:    "invalid-email",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid email, got %d", w.Code)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupTestRouter()
	router.POST("/api/auth/register", Register)

	body := models.RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for short password, got %d", w.Code)
	}
}

func TestLogin_Success(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create a test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := models.User{
		ID:        primitive.NewObjectID(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Role:      models.RoleUser,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	database.UsersCollection.InsertOne(ctx, testUser)

	router := setupTestRouter()
	router.POST("/api/auth/login", Login)

	body := models.LoginInput{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	data := response["data"].(map[string]interface{})
	if data["token"] == nil {
		t.Error("Expected token in response")
	}

	// Verify password is not returned
	user := data["user"].(map[string]interface{})
	if user["password"] != nil {
		t.Error("Password should not be returned in response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	// Create a test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	testUser := models.User{
		ID:        primitive.NewObjectID(),
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Role:      models.RoleUser,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	database.UsersCollection.InsertOne(ctx, testUser)

	router := setupTestRouter()
	router.POST("/api/auth/login", Login)

	body := models.LoginInput{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	router := setupTestRouter()
	router.POST("/api/auth/login", Login)

	body := models.LoginInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
