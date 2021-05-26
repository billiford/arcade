package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetToken returns a new access token for a given provider.
func (ctl *Controller) GetToken(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		provider = "google"
	}

	tokenizer, ok := ctl.Tokenizers[provider]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Unsupported token provider: %s", provider)})

		return
	}

	t, err := tokenizer.Token(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})

		return
	}

	c.JSON(http.StatusOK, gin.H{"token": t})
}
