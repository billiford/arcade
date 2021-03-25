package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.homedepot.com/cd/arcade/pkg/google"
	"github.homedepot.com/cd/arcade/pkg/rancher"
)

func NewApiKeyAuth(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Header.Get("Api-Key") != apiKey {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "bad api key"})
			return
		}

		c.Next()
	}
}

func SetGoogleClient(g google.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(google.Key, g)
		c.Next()
	}
}

func SetRancherClient(r rancher.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(rancher.Key, r)
		c.Next()
	}
}
