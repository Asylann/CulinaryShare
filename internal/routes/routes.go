package routes

import (
	"github.com/culinaryshare/backend/internal/handlers"
	"github.com/culinaryshare/backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "message": "CulinaryShare API is running"})
	})

	// API v1 group
	api := router.Group("/api")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register) // 1. POST /api/auth/register
			auth.POST("/login", handlers.Login)       // 2. POST /api/auth/login
		}

		// User routes (protected)
		users := api.Group("/users")
		users.Use(middleware.AuthRequired())
		{
			users.GET("/me", handlers.GetMe) // 3. GET /api/users/me
		}

		// Category routes
		categories := api.Group("/categories")
		{
			categories.GET("", handlers.GetCategories) // 4. GET /api/categories

			// Admin only routes
			adminCategories := categories.Group("")
			adminCategories.Use(middleware.AuthRequired(), middleware.AdminRequired())
			{
				adminCategories.POST("", handlers.CreateCategory) // 5. POST /api/categories (ADMIN)
			}
		}

		// Recipe routes
		recipes := api.Group("/recipes")
		{
			recipes.GET("", handlers.GetRecipes)    // 6. GET /api/recipes (filters + pagination)
			recipes.GET("/:id", handlers.GetRecipe) // 8. GET /api/recipes/:id

			// Protected recipe routes
			protectedRecipes := recipes.Group("")
			protectedRecipes.Use(middleware.AuthRequired())
			{
				protectedRecipes.POST("", handlers.CreateRecipe)                         // 7. POST /api/recipes
				protectedRecipes.PUT("/:id", handlers.UpdateRecipe)                      // 9. PUT /api/recipes/:id
				protectedRecipes.PATCH("/:id/ingredients", handlers.UpdateIngredients)   // 10. PATCH /api/recipes/:id/ingredients
				protectedRecipes.POST("/:id/reviews", handlers.AddReview)                // 11. POST /api/recipes/:id/reviews
				protectedRecipes.DELETE("/:id/reviews/:reviewId", handlers.DeleteReview) // 12. DELETE /api/recipes/:id/reviews/:reviewId
			}
		}

		// Analytics routes (public)
		analytics := api.Group("/analytics")
		{
			analytics.GET("/top-rated", handlers.GetTopRatedRecipes) // 13. GET /api/analytics/top-rated
		}
	}
}
