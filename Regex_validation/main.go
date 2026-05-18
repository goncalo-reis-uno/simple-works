package main

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	router.GET("/email", func(c *gin.Context) {
		email := c.Query("email")
		if email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
			return
		}

		// Validate email format
		if !isValidEmail(email) {
			c.JSON(http.StatusBadRequest, gin.H{
				"output": "Email is invalid",
				"valid": false,
				"received_email": email})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"output": "Email is valid",
			"valid": true,
			"received_email": email})
	})

	router.Run(":8080")
}

func isValidEmail(email string) bool {
	// regex done on task
	re := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(re, email)
	return matched
}

//Use invoke-restmethod "http://localhost:8080/email?email=exemplo@exemplo.com"| curl "http://localhost:8080/email?email=exemplo@exemplo.com" for testing
//Alternatively you can use npx vitest to run test cases in the test file
