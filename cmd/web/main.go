// in main.go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Tell Gin where to find static files (CSS, images)
	router.Static("/static", "./static")
	// Tell Gin where to find HTML template files
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		// In a real app, you would get this data from your database
		sinsData := []map[string]interface{}{
			{"Description": "Used a 'temporary' fix that is now 2 years old.", "Count": 5},
			{"Description": "Wrote code without any tests.", "Count": 12},
		}

		// Render the index.html template, passing the data to it
		c.HTML(http.StatusOK, "index.html", gin.H{
			"sins": sinsData,
		})
	})

	router.Run(":8080")
}
