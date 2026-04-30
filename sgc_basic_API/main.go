package main

import (
	"log"
	"net/http"
	"time"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)


// models
type User struct {
	ID           string    `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Email        string    `json:"email" binding:"required,email" gorm:"not null;unique"`
	Password     string    `json:"-" binding:"required" gorm:"not null"`
	Name         string    `json:"name" binding:"required" gorm:"not null"`
	Role         string    `json:"role" gorm:"default:USER"`
	RefreshToken string    `json:"-" gorm:"type:text"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
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

	if err := DB.AutoMigrate(&User{}, &Contact{}); err != nil {
		log.Fatal("Migration failed:", err)
	}

	fmt.Printf("Database connection established on port %d\n", 7777)
}


//Password hashing
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(hashedPassword), err
}

//Password verification
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

//JWT access token generation
func GenerateJWT(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":  time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) // The token is signed here
	return token.SignedString([]byte("your_secret_key")) // Returns the signed token
}


// Main function
func main() {
	InitDB() // database initialization


	r := gin.Default() // router initialization


	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

