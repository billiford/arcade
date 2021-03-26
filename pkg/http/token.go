package http

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/homedepot/arcade/pkg/google"
	"github.com/homedepot/arcade/pkg/rancher"
)

var (
	err             error
	googleMux       sync.Mutex
	t               time.Time
	token           string
	expiration      = 1 * time.Minute
	rancherMux      sync.Mutex
	kubeconfigToken rancher.KubeconfigToken
)

func GetToken(c *gin.Context) {
	provider := c.Query("provider")

	switch provider {
	case "rancher":
		getRancherToken(c)
	case "google", "":
		getGoogleToken(c)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported token provider: %s", provider)})
		return
	}
}

func getGoogleToken(c *gin.Context) {
	googleMux.Lock()
	defer googleMux.Unlock()

	if time.Since(t) > expiration || token == "" {
		googleClient := google.Instance(c)

		token, err = googleClient.NewToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		t = time.Now().In(time.UTC)
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func getRancherToken(c *gin.Context) {
	if time.Now().In(time.UTC).After(kubeconfigToken.ExpiresAt) || kubeconfigToken.Token == "" {
		rancherMux.Lock()
		defer rancherMux.Unlock()

		rancherClient := rancher.Instance(c)
		if rancherClient == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token provider not configured: rancher"})
			return
		}

		kubeconfigToken, err = rancherClient.NewToken(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"token": kubeconfigToken.Token})
}
