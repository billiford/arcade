package http

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/homedepot/arcade/pkg/google"
	"github.com/homedepot/arcade/pkg/rancher"
	"golang.org/x/oauth2"
)

var (
	err              error
	googleMux        sync.Mutex
	googleExpiration time.Time
	googleToken      *oauth2.Token
	rancherMux       sync.Mutex
	kubeconfigToken  rancher.KubeconfigToken
)

// GetToken returns a new access token for a given provider.
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

	if time.Now().UTC().After(googleExpiration) || googleToken == nil {
		googleClient := google.Instance(c)

		googleToken, err = googleClient.NewToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Set the expiration for the Google token to be 90% expiry-threshold.
		googleExpiration = time.Now().UTC().Add((time.Until(googleToken.Expiry) / 10) * 9)
	}

	c.JSON(http.StatusOK, gin.H{"token": googleToken.AccessToken})
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
