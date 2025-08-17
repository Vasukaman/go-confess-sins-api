package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Serve static files (like script.js and style.css) from the "/static" path
	router.Static("/static", "./web/static")

	// Serve the main index.html file for the root path
	router.StaticFile("/", "./web/templates/index.html")

	router.Run(":9090")
}
