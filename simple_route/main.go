package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.GET("/date", func(c *gin.Context) {
		currentDate := time.Now().Format("02-01-2006")

		c.JSON(http.StatusOK, gin.H{
			"current_date": currentDate,
		})
	})

	r.Run(":8080")
}
