package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.homedepot.com/cd/arcade/pkg/http"
	"github.homedepot.com/cd/arcade/pkg/middleware"
)

var (
	r = gin.Default()
)

func init() {
	apiKey := os.Getenv("ARCADE_API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY not set; exiting.")
	}

	gin.ForceConsoleColor()

	r.Use(middleware.NewApiKeyAuth(apiKey))

	r.GET("/tokens", http.GetToken)
}

// Run arcade on port 1982.
func main() {
	r.Run(":1982")
}
