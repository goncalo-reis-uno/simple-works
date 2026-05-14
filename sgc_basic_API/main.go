package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// models
type User struct {
	ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email     string    `json:"email" binding:"required,email" gorm:"not null;unique"`
	Password  string    `json:"password" binding:"required" gorm:"not null"`
	Name      string    `json:"name" binding:"required" gorm:"not null"`
	Role      string    `json:"role" gorm:"default:user"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type RefreshToken struct {
	ID        string    `json:"id" gorm:"type:uuid;primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Token     string    `json:"token" gorm:"type:text;not null;unique"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
}

type Contact struct {
	ID        string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:uuid;not null;index"`
	Name      string    `json:"name" binding:"required" gorm:"not null"`
	Email     string    `json:"email" binding:"required,email" gorm:"not null;index"`
	Phone     string    `json:"phone" binding:"required" gorm:"not null"`
	Address   string    `json:"address" binding:"required" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// database connection
var DB *gorm.DB

func InitDB() {
	dsn := "host=localhost user=postgres password=admin dbname=sgc_basic_API port=7777 sslmode=disable TimeZone=UTC"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	DB = db

	if err := DB.AutoMigrate(&User{}, &Contact{}, &RefreshToken{}); err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Printf("Database connection established on port %d\n", 7777)
}

// Password hashing
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(hashedPassword), err
}

// Password verification
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Password validation function
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	hasUpper := false
	hasNumber := false
	for _, c := range password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		}
		if c >= '0' && c <= '9' {
			hasNumber = true
		}
	}
	if !hasUpper {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasNumber {
		return fmt.Errorf("password must contain at least one number")
	}
	return nil
}

// JWT token generation
var jwtKey = []byte("secret")
var refreshKey = []byte("refresh_secret")

func GenerateJWT(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // The token is signed here
	return token.SignedString(jwtKey)                          // Returns the signed token
}

func GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // Refresh token expires in 7 days
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(refreshKey)
}

// Middleware for JWT authentication
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the token from the Authorization header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			c.Abort()
			return
		}

		// Parse the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		// Validate the token
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		// Set user information in context
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			c.Abort()
			return
		}
		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])
		c.Next()
	}
}

// Is Admin middleware
func IsAdmin(c *gin.Context) bool {
	role, exists := c.Get("role")
	return exists && role == "admin"
}

// Error handling middleware
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 && !c.Writer.Written() {
			err := c.Errors.Last()

			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
		}
	}
}

// Main function
func main() {
	InitDB() // database initialization

	r := gin.Default() // router initialization

	r.Use(ErrorHandlingMiddleware()) // global error handling middleware

	// routes

	// ping route for testing
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// User registration route
	r.POST("/register", func(c *gin.Context) {
		var user User
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user.ID = uuid.New().String()

		// Validate password
		if err := ValidatePassword(user.Password); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			c.Error(err)
			return
		}

		user.Password = hashedPassword

		if err := DB.Create(&user).Error; err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	})

	// User login route
	r.POST("/login", func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}

		var user User

		// Check user email
		if err := DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// Check user password
		if !CheckPasswordHash(input.Password, user.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		// Generate JWT Access Token
		accessToken, err := GenerateJWT(user.ID, user.Role)
		if err != nil {
			c.Error(err)
			return
		}

		// Generate JWT Refresh Token
		refreshToken, err := GenerateRefreshToken(user.ID)
		if err != nil {
			c.Error(err)
			return
		}

		refreshTokenRecord := RefreshToken{
			ID:        uuid.New().String(),
			UserID:    user.ID,
			Token:     refreshToken,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		}

		// Save refresh token to database
		if err := DB.Create(&refreshTokenRecord).Error; err != nil {
			c.Error(err)
			return
		}

		// Return access and refresh tokens
		c.JSON(http.StatusOK, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
		})
	})

	// Refresh token route
	r.POST("/refresh", func(c *gin.Context) {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}

		// Check if refresh token is provided
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		// Parse the refresh token
		token, err := jwt.Parse(body.RefreshToken, func(token *jwt.Token) (interface{}, error) {
			return refreshKey, nil
		})

		// Check if refresh token is valid
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userID := claims["user_id"].(string)

		var storedToken RefreshToken

		// Check if the refresh token exists in the database
		if err := DB.Where("token = ?", body.RefreshToken).First(&storedToken).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
			return
		}

		// Check if the refresh token has expired
		if storedToken.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token expired"})
			return
		}

		// Check if the refresh token belongs to the user
		if storedToken.UserID != userID {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token does not match user"})
			return
		}

		var user User

		// Check if the user exists
		if err := DB.First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		// Generate new access token
		newAccessToken, err := GenerateJWT(user.ID, user.Role)
		if err != nil {
			c.Error(err)
			return
		}
		// Return new access token
		c.JSON(http.StatusOK, gin.H{
			"access_token": newAccessToken,
		})
	})

	// Logout route
	r.POST("/logout", func(c *gin.Context) {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}

		// Check if refresh token is provided
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}

		var storedToken RefreshToken

		// Check if the refresh token exists in the database
		if err := DB.Where("token = ?", body.RefreshToken).First(&storedToken).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token not found"})
			return
		}

		// Delete the refresh token from the database to log out the user
		if err := DB.Delete(&storedToken).Error; err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
	})

	// Get all users route (admin only)
	r.GET("/users", AuthMiddleware(), func(c *gin.Context) {
		if !IsAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

		var users []User

		// Fetch all users from the database and checks for errors
		if err := DB.Find(&users).Error; err != nil {
			c.Error(err)
			return
		}
		// Return the list of users
		c.JSON(http.StatusOK, users)
	})

	// Delete user route (admin only)
	r.DELETE("/users/:id", AuthMiddleware(), func(c *gin.Context) {
		if !IsAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		userID := c.Param("id")

		// Delete the user from the database and checks for errors
		if err := DB.Delete(&User{}, "id = ?", userID).Error; err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user deleted successfully"})
	})

	// Update any user credentials route (admin only)
	r.PUT("/users/:id", AuthMiddleware(), func(c *gin.Context) {
		if !IsAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		// get user ID from URL
		userID := c.Param("id")

		var user User

		// checks if the user exists in the database
		if err := DB.First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		// input should follow this structure for updating user credentials
		var input struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		user.Name = input.Name
		user.Email = input.Email

		// update password if provided
		if input.Password != "" {
			if err := ValidatePassword(input.Password); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			hashedPassword, err := HashPassword(input.Password)
			if err != nil {
				c.Error(err)
				return
			}
			user.Password = hashedPassword
		}
		// save updated user to the database
		if err := DB.Save(&user).Error; err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
	})

	// Update user credentials route (self)
	r.PUT("/users/me", AuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var user User

		// Fetch the user from the database and checks for errors
		if err := DB.First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
			return
		}

		var input struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update fields
		user.Name = input.Name
		user.Email = input.Email

		// Update password if provided
		if input.Password != "" {
			// Validate password
			if err := ValidatePassword(input.Password); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			hashedPassword, err := HashPassword(input.Password)
			if err != nil {
				c.Error(err)
				return
			}
			user.Password = hashedPassword
		}

		// Save the updated user to the database
		if err := DB.Save(&user).Error; err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "user updated successfully"})
	})

	// Contact routes
	// Create contact route
	r.POST("/contacts", AuthMiddleware(), func(c *gin.Context) {
		var contact Contact

		if err := c.ShouldBindJSON(&contact); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Generate UUID for contact
		contact.ID = uuid.New().String()

		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		// Assign user ID to contact
		contact.UserID = userID.(string)

		if err := DB.Create(&contact).Error; err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "contact created successfully"})
	})

	// Get all contacts route
	r.GET("/contacts", AuthMiddleware(), func(c *gin.Context) {
		// Get user ID from context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
			return
		}

		var contacts []Contact

		query := DB.Model(&Contact{})

		// Fetch contacts based on user role
		if !IsAdmin(c) {
			query = query.Where("user_id = ?", userID)
		}

		// Filtering
		name := c.Query("name")
		email := c.Query("email")

		// Verify name and email query (ILIKE for case-insensitive search and % for partial matching)
		if name != "" {
			query = query.Where("name ILIKE ?", "%"+name+"%")
		}

		if email != "" {
			query = query.Where("email ILIKE ?", "%"+email+"%")
		}

		// Sorting
		sortBy := c.DefaultQuery("sort", "created_at")
		order := c.DefaultQuery("order", "desc")

		query = query.Order(sortBy + " " + order)

		if err := query.Find(&contacts).Error; err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, contacts)
	})

	// Get contact by ID route
	r.GET("/contacts/:id", AuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		contactID := c.Param("id")

		var contact Contact
		query := DB.Where("id = ?", contactID)

		if !IsAdmin(c) {
			query = query.Where("user_id = ?", userID)
		}

		if err := query.First(&contact).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
			return
		}

		c.JSON(http.StatusOK, contact)
	})

	// Update contact route
	r.PUT("/contacts/:id", AuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		contactID := c.Param("id")

		var contact Contact
		query := DB.Where("id = ?", contactID)

		if !IsAdmin(c) {
			query = query.Where("user_id = ?", userID)
		}

		if err := query.First(&contact).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "contact not found"})
			return
		}

		var input Contact
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		contact.Name = input.Name
		contact.Email = input.Email
		contact.Phone = input.Phone
		contact.Address = input.Address

		if err := DB.Save(&contact).Error; err != nil {
			c.Error(err)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "contact updated successfully"})
	})

	// Delete contact route
	r.DELETE("/contacts/:id", AuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		contactID := c.Param("id")

		query := DB.Where("id = ?", contactID)

		// Normal users can only delete own contacts
		if !IsAdmin(c) {
			query = query.Where("user_id = ?", userID)
		}

		// Delete the contact and checks for errors
		result := query.Delete(&Contact{})

		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "contact not found or not authorized"})
			return
		}

		if result.Error != nil {
			c.Error(result.Error)
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "contact deleted successfully"})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
