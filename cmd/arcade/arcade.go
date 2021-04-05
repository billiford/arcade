package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/homedepot/arcade/pkg/google"
	arcadehttp "github.com/homedepot/arcade/pkg/http"
	"github.com/homedepot/arcade/pkg/middleware"
	"github.com/homedepot/arcade/pkg/rancher"
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

	r.GET("/tokens", arcadehttp.GetToken)
}

func mustGetenv(env string) (s string) {
	if s = os.Getenv(env); s == "" {
		log.Fatal(env + " not set; exiting.")
	}

	return
}

func mustInstantiateRancherClient() rancher.Client {
	url := mustGetenv("RANCHER_URL")
	username := mustGetenv("RANCHER_USERNAME")
	password := mustGetenv("RANCHER_PASSWORD")

	rancherClient := rancher.NewClient()
	rancherClient.WithURL(url)
	rancherClient.WithUsername(username)
	rancherClient.WithPassword(password)

	if caCerts := os.Getenv("RANCHER_CACERTS"); caCerts != "" {
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		rootCAs.AppendCertsFromPEM([]byte(caCerts))

		t := &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rootCAs,
			},
		}

		rancherClient.WithTransport(t)
	}

	return rancherClient
}

// Run arcade on port 1982.
func main() {
	err := r.Run(":1982")
	if err != nil {
		panic(err)
	}
}
