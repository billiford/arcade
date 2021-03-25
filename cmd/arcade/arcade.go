package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.homedepot.com/cd/arcade/pkg/google"
	"github.homedepot.com/cd/arcade/pkg/http"
	"github.homedepot.com/cd/arcade/pkg/middleware"
	"github.homedepot.com/cd/arcade/pkg/rancher"
)

var (
	r = gin.Default()
)

func init() {
	gin.ForceConsoleColor()

	apiKey := mustGetenv("ARCADE_API_KEY")
	r.Use(middleware.NewApiKeyAuth(apiKey))

	googleClient := google.NewClient()
	r.Use(middleware.SetGoogleClient(googleClient))

	if s := os.Getenv("RANCHER_ENABLED"); s == "TRUE" {
		rancherClient := mustInstantiateRancherClient()
		r.Use(middleware.SetRancherClient(rancherClient))
	}

	r.GET("/tokens", http.GetToken)
}

func mustGetenv(env string) (s string) {
	if s = os.Getenv(env); s == "" {
		log.Fatal(env + " not set; exiting.")
	}

	return
}

func mustInstantiateRancherClient() rancher.Client {
	rancherURL := mustGetenv("RANCHER_URL")
	rancherUsername := mustGetenv("RANCHER_USERNAME")
	rancherPassword := mustGetenv("RANCHER_PASSWORD")

	rancherClient := rancher.NewClient()
	rancherClient.WithURL(rancherURL)
	rancherClient.WithUsername(rancherUsername)
	rancherClient.WithPassword(rancherPassword)

	return rancherClient
}

// Run arcade on port 1982.
func main() {
	err := r.Run(":1982")
	if err != nil {
		panic(err)
	}
}
