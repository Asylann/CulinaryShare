package handlers

import (
	"context"
	"time"

	"github.com/culinaryshare/backend/internal/config"
	"github.com/culinaryshare/backend/internal/database"
	"github.com/culinaryshare/backend/internal/models"
	"github.com/culinaryshare/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// Register handles user registration
func Register(c *gin.Context) {
	var input models.RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	// Validate input
	validationErrors := &utils.ValidationErrors{}
	validationErrors.Add(utils.ValidateRequired(input.Username, "username"))
	validationErrors.Add(utils.ValidateEmail(input.Email))
	validationErrors.Add(utils.ValidatePassword(input.Password))

	if validationErrors.HasErrors() {
		utils.BadRequest(c, validationErrors.Error(), nil)
		return
	}

	// Hash password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.InternalServerError(c, "Failed to hash password", err)
		return
	}

	// Create user
	now := time.Now()
	user := models.User{
		ID:        primitive.NewObjectID(),
		Username:  input.Username,
		Email:     input.Email,
		Password:  string(hashedPassword), // Stored hashed, never returned
		Role:      models.RoleUser,
		CreatedAt: now,
		UpdatedAt: now,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = database.UsersCollection.InsertOne(ctx, user)
	if err != nil {
		// Check for duplicate key error (email or username already exists)
		if isDuplicateKeyError(err) {
			utils.Conflict(c, "Email or username already exists")
			return
		}
		utils.InternalServerError(c, "Failed to create user", err)
		return
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		utils.InternalServerError(c, "Failed to generate token", err)
		return
	}

	utils.Created(c, "User registered successfully", models.LoginResponse{
		Token: token,
		User:  user.ToResponse(), // Password is excluded via ToResponse()
	})
}

// Login handles user login
func Login(c *gin.Context) {
	var input models.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.BadRequest(c, "Invalid request body", err)
		return
	}

	// Validate input
	validationErrors := &utils.ValidationErrors{}
	validationErrors.Add(utils.ValidateEmail(input.Email))
	validationErrors.Add(utils.ValidateRequired(input.Password, "password"))

	if validationErrors.HasErrors() {
		utils.BadRequest(c, validationErrors.Error(), nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find user by email
	var user models.User
	err := database.UsersCollection.FindOne(ctx, bson.M{"email": input.Email}).Decode(&user)
	if err != nil {
		utils.Unauthorized(c, "Invalid email or password")
		return
	}

	// Compare password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		utils.Unauthorized(c, "Invalid email or password")
		return
	}

	// Generate JWT token
	token, err := generateToken(user)
	if err != nil {
		utils.InternalServerError(c, "Failed to generate token", err)
		return
	}

	utils.OK(c, "Login successful", models.LoginResponse{
		Token: token,
		User:  user.ToResponse(), // Password is excluded via ToResponse()
	})
}

// generateToken creates a JWT token for the user
func generateToken(user models.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(config.AppConfig.JWTExpiresIn) * time.Second)

	claims := jwt.MapClaims{
		"userId":   user.ID.Hex(),
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"exp":      expirationTime.Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

// isDuplicateKeyError checks if the error is a MongoDB duplicate key error
func isDuplicateKeyError(err error) bool {
	return err != nil && (contains(err.Error(), "E11000") || contains(err.Error(), "duplicate key"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
