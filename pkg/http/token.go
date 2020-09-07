package http

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.homedepot.com/cd/arcade/pkg/google"
)

var (
	mux        sync.Mutex
	t          time.Time
	token      string
	err        error
	expiration = 1 * time.Minute
)

func GetToken(c *gin.Context) {
	mux.Lock()
	defer mux.Unlock()

	if time.Since(t) > expiration || token == "" {
		token, err = google.NewToken()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		t = time.Now().In(time.UTC)
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
