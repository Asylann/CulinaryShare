# CulinaryShare

A full-stack recipe management web application built with Go, Gin framework, MongoDB Atlas, and vanilla HTML/CSS/JavaScript for the NoSQL Final Exam project.

## 👥 Team Information

| Name | Contribution |
|------|------------|
| Nurassyl  Orazbek | Auth, User handlers, JWT middleware, Tests, Frontend Auth |
| Asylan Usen | Recipe, Category, Analytics handlers, Documentation, Frontend UI |

## 📋 Project Overview

CulinaryShare is a complete web application for managing recipes, ingredients, reviews, and users. It features:
- **Backend**: RESTful API with JWT-based authentication and role-based access control (USER/ADMIN)
- **Frontend**: Modern responsive UI with dark theme, glassmorphism, and smooth animations
- **Database**: MongoDB with embedded and referenced documents, aggregation pipelines, and compound indexes
- CRUD operations for recipes and categories
- Embedded documents (ingredients, reviews) and referenced documents (userId, categoryId)
- Multi-stage aggregation pipelines for analytics
- Atomic rating updates

## 🏗️ Architecture

```
/cmd
  └── main.go           # Application entry point

/internal
  /config
    └── config.go       # Environment configuration
  
  /database
    └── mongo.go        # MongoDB connection + index creation
  
  /models
    ├── user.go         # User model
    ├── recipe.go       # Recipe model (with embedded Ingredient, Review)
    └── category.go     # Category model
  
  /handlers
    ├── auth.go         # Register, Login (bcrypt + JWT)
    ├── user.go         # User profile
    ├── recipe.go       # Recipe CRUD, ingredients, reviews
    ├── category.go     # Category CRUD
    └── analytics.go    # Aggregation pipelines
  
  /middleware
    └── auth.go         # JWT validation + role checking
  
  /routes
    └── routes.go       # API route definitions

/frontend
  ├── index.html        # Landing page with hero, top rated, categories
  ├── login.html        # User login page
  ├── register.html     # User registration page  
  ├── recipes.html      # All recipes with filters and pagination
  ├── recipe.html       # Recipe detail with reviews
  ├── create-recipe.html # Create/edit recipe form
  ├── categories.html   # Browse all categories
  ├── profile.html      # User profile and settings
  /css
    └── styles.css      # Complete design system
  /js
    ├── api.js          # API client with auth management
    └── utils.js        # Utility functions

/utils
  ├── response.go       # Centralized error responses
  └── validation.go     # Input validation
```

## 🎨 Frontend Pages (6 pages)

| Page | Description | Features |
|------|-------------|----------|
| **index.html** | Landing page | Hero section, top rated recipes, categories preview, recent recipes |
| **login.html** | Login page | Form validation, JWT token storage |
| **register.html** | Registration page | Form validation, auto-login after registration |
| **recipes.html** | All recipes | Search, category filter, sort, pagination |
| **recipe.html** | Recipe details | Ingredients checklist, reviews, owner actions |
| **create-recipe.html** | Create/Edit recipe | Dynamic ingredients, form validation |
| **categories.html** | Categories | Grid view, admin can add categories |
| **profile.html** | User profile | Stats, my recipes, settings |

## 📦 MongoDB Schema Design

### Embedded Documents

**Ingredients** (in Recipe):
```json
{
  "name": "Spaghetti",
  "quantity": "400",
  "unit": "g"
}
```
- **Why embedded:** Ingredients are always accessed with recipes, small array, no need for separate queries.

**Reviews** (in Recipe):
```json
{
  "_id": ObjectId,
  "userId": ObjectId,
  "username": "johndoe",
  "rating": 5,
  "comment": "Delicious!",
  "createdAt": ISODate
}
```
- **Why embedded:** Common read pattern (show recipe with reviews), denormalized for performance.

### Referenced Documents

**userId** and **categoryId** in Recipe:
```json
{
  "userId": ObjectId("..."),
  "categoryId": ObjectId("...")
}
```
- **Why referenced:** Users and categories are separate entities that can exist independently and be shared across recipes.

## 🔍 MongoDB Indexes

### 1. Unique Index on `users.email`
```javascript
db.users.createIndex({ email: 1 }, { unique: true })
```
Ensures no duplicate email addresses in the system.

### 2. Unique Index on `users.username`
```javascript
db.users.createIndex({ username: 1 }, { unique: true })
```
Ensures unique usernames.

### 3. Compound Index on `recipes(categoryId, averageRating)`
```javascript
db.recipes.createIndex({ categoryId: 1, averageRating: -1 })
```
**Purpose:** Optimizes the common query pattern "show best recipes in category X". This index supports:
- Filtering by categoryId (equality match)
- Sorting by averageRating descending
- Combined filter + sort in a single index scan

### 4. Index on `recipes.userId`
```javascript
db.recipes.createIndex({ userId: 1 })
```
Optimizes queries filtering recipes by author.

## 📊 Aggregation Pipeline

The `/api/analytics/top-rated` endpoint uses a multi-stage aggregation pipeline:

```javascript
[
  { $match: { reviewCount: { $gt: 0 } } },           // Stage 1: Filter recipes with reviews
  { $lookup: { from: "categories", ... } },          // Stage 2: Join categories
  { $lookup: { from: "users", ... } },               // Stage 3: Join users
  { $unwind: "$category" },                          // Stage 4: Flatten category
  { $unwind: "$author" },                            // Stage 5: Flatten author
  { $project: { ... } },                             // Stage 6: Select fields
  { $sort: { averageRating: -1, reviewCount: -1 } }, // Stage 7: Sort
  { $limit: 10 }                                     // Stage 8: Limit results
]
```

## 🔧 MongoDB Operations Used

| Operation | Usage |
|-----------|-------|
| `$push` | Adding ingredients/reviews to arrays |
| `$pull` | Removing ingredients/reviews from arrays |
| `$inc` | Incrementing review count |
| `$set` | Updating specific fields |
| Positional updates | Updating/removing specific array elements |
| Aggregation | Multi-stage pipeline for analytics |

## 🚀 How to Run

### Prerequisites
- Go 1.21+
- MongoDB Atlas account

### 1. Clone and Setup
```bash
git clone <repository>
cd NOSQL_Final
```

### 2. Configure Environment
```bash
cp .env.example .env
# Edit .env with your MongoDB Atlas URI
```

### 3. Install Dependencies
```bash
go mod tidy
```

### 4. Run the Backend
```bash
go run ./cmd/main.go
```

### 5. Run the Frontend
```bash
# Option 1: Using Python (recommended)
cd frontend && python3 -m http.server 3000

# Option 2: Using the serve script
./frontend/serve.sh

# Option 3: Using npx
npx serve frontend

# Option 4: Open directly in browser
# Just open frontend/index.html in your browser
```

The frontend will be available at: http://localhost:3000

### 6. Run Tests
```bash
go test ./internal/handlers/... -v
```

### Using Docker (Recommended - One Command!)
```bash
# Run the entire project (backend + frontend) with one command:
docker-compose up --build
```

This will start:
- **Backend API** at http://localhost:8080
- **Frontend** at http://localhost:3000

The frontend uses nginx to serve static files and proxy API requests to the backend.

## 🔌 Connecting to MongoDB Atlas

1. Create a cluster on [MongoDB Atlas](https://www.mongodb.com/atlas)
2. Create a database user
3. Whitelist your IP address
4. Get the connection string
5. Set `MONGODB_ATLAS_URI` in your `.env` file:
   ```
   MONGODB_ATLAS_URI=mongodb+srv://<username>:<password>@cluster.mongodb.net/culinaryshare?retryWrites=true&w=majority
   ```

## 📡 API Endpoints

| # | Method | Endpoint | Auth | Description |
|---|--------|----------|------|-------------|
| 1 | POST | `/api/auth/register` | - | Register new user |
| 2 | POST | `/api/auth/login` | - | Login, get JWT |
| 3 | GET | `/api/users/me` | ✓ | Get current user |
| 4 | GET | `/api/categories` | - | List categories |
| 5 | POST | `/api/categories` | ADMIN | Create category |
| 6 | GET | `/api/recipes` | - | List recipes (filter + pagination) |
| 7 | POST | `/api/recipes` | ✓ | Create recipe |
| 8 | GET | `/api/recipes/:id` | - | Get recipe by ID |
| 9 | PUT | `/api/recipes/:id` | Owner/ADMIN | Update recipe |
| 10 | PATCH | `/api/recipes/:id/ingredients` | Owner/ADMIN | Add/remove ingredients |
| 11 | POST | `/api/recipes/:id/reviews` | ✓ | Add review |
| 12 | DELETE | `/api/recipes/:id/reviews/:reviewId` | Owner/ADMIN | Delete review |
| 13 | GET | `/api/analytics/top-rated` | - | Get top-rated recipes |

## 📁 Additional Files

- `docs/swagger.yaml` - OpenAPI 3.0 specification
- `docs/postman_collection.json` - Postman collection
- `seed_data.json` - Sample data for testing

## 📝 License

This project is for educational purposes only - NoSQL Final Exam.
