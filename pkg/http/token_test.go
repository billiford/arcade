package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/homedepot/arcade/pkg/google/googlefakes"
	arcadehttp "github.com/homedepot/arcade/pkg/http"
	"github.com/homedepot/arcade/pkg/middleware"
	"github.com/homedepot/arcade/pkg/rancher"
	"github.com/homedepot/arcade/pkg/rancher/rancherfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"
)

type Tokens struct {
	Token string `json:"token"`
	Error string `json:"error"`
}

var (
	err               error
	svr               *httptest.Server
	uri               string
	req               *http.Request
	body              *bytes.Buffer
	res               *http.Response
	tokens            Tokens
	fakeGoogleClient  *googlefakes.FakeClient
	fakeRancherClient *rancherfakes.FakeClient
	fakeGoogleToken   = &oauth2.Token{
		AccessToken: "fake-google-token",
	}
	fakeRancherToken = rancher.KubeconfigToken{
		Token: "fake-rancher-token",
	}
	expiredRancherToken = rancher.KubeconfigToken{
		ExpiresAt: time.Now().In(time.UTC).Add(-24 * time.Hour),
		Token:     "expired-rancher-token",
	}
	validRancherToken = rancher.KubeconfigToken{
		ExpiresAt: time.Now().In(time.UTC).Add(1 * time.Hour),
		Token:     "valid-rancher-token",
	}
)

var _ = Describe("Token", func() {
	BeforeEach(func() {
		// Disable debug logging.
		gin.SetMode(gin.ReleaseMode)
	})

	Describe("#GetToken", func() {
		When("provider is not supported", func() {
			BeforeEach(func() {
				r := gin.New()
				r.Use(gin.Recovery())
				r.GET("/tokens", arcadehttp.GetToken)

				svr = httptest.NewServer(r)
				uri = svr.URL + "/tokens?provider=fake"
				body = &bytes.Buffer{}
			})

			AfterEach(func() {
				svr.Close()
				res.Body.Close()
			})

			JustBeforeEach(func() {
				req, _ = http.NewRequest(http.MethodGet, uri, nil)
				res, err = http.DefaultClient.Do(req)
			})

			It("succeeds", func() {
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				b, _ := ioutil.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("Unsupported token provider: fake"))
			})
		})
	})

	Describe("#GetGoogleToken", func() {
		BeforeEach(func() {
			fakeGoogleClient = &googlefakes.FakeClient{}
			fakeGoogleClient.NewTokenReturns(fakeGoogleToken, nil)

			// Create new gin instead of using gin.Default().
			// This disables request logging which we don't want for tests.
			r := gin.New()
			r.Use(gin.Recovery())
			r.Use(middleware.SetGoogleClient(fakeGoogleClient))
			r.GET("/tokens", arcadehttp.GetToken)

			svr = httptest.NewServer(r)
			uri = svr.URL + "/tokens?provider=google"
			body = &bytes.Buffer{}
		})

		AfterEach(func() {
			svr.Close()
			res.Body.Close()
		})

		JustBeforeEach(func() {
			req, _ = http.NewRequest(http.MethodGet, uri, nil)
			res, err = http.DefaultClient.Do(req)
		})

		When("getting a new token from google fails", func() {
			BeforeEach(func() {
				fakeGoogleClient.NewTokenReturns(nil, errors.New("error getting token from google"))
			})

			It("returns an internal server error", func() {
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
				b, _ := ioutil.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("error getting token from google"))
			})
		})

		When("no provider is specified", func() {
			BeforeEach(func() {
				uri = svr.URL + "/tokens"
			})

			It("defaults to google", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				b, _ := ioutil.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Token).To(Equal("fake-google-token"))
			})
		})

		When("it succeeds", func() {
			It("succeeds", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				b, _ := ioutil.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Token).To(Equal("fake-google-token"))
			})
		})
	})

	Describe("#GetRancherToken", func() {
		BeforeEach(func() {
			fakeRancherClient = &rancherfakes.FakeClient{}
			fakeRancherClient.NewTokenReturns(validRancherToken, nil)

			// Create new gin instead of using gin.Default().
			// This disables request logging which we don't want for tests.
			r := gin.New()
			r.Use(gin.Recovery())
			r.Use(middleware.SetRancherClient(fakeRancherClient))
			r.GET("/tokens", arcadehttp.GetToken)

			svr = httptest.NewServer(r)
			uri = svr.URL + "/tokens?provider=rancher"
			body = &bytes.Buffer{}
		})

		AfterEach(func() {
			svr.Close()
			res.Body.Close()
		})

		JustBeforeEach(func() {
			req, _ = http.NewRequest(http.MethodGet, uri, nil)
			res, err = http.DefaultClient.Do(req)
		})

		When("rancher is not a configured provider", func() {
			BeforeEach(func() {
				r := gin.New()
				r.Use(gin.Recovery())
				r.GET("/tokens", arcadehttp.GetToken)

				svr = httptest.NewServer(r)
				uri = svr.URL + "/tokens?provider=rancher"
				body = &bytes.Buffer{}
			})

			It("returns a bad request error", func() {
				Expect(res.StatusCode).To(Equal(http.StatusBadRequest))
				b, _ := ioutil.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("token provider not configured: rancher"))
			})
		})

		When("getting a new token from rancher fails", func() {
			BeforeEach(func() {
				fakeRancherClient.NewTokenReturns(rancher.KubeconfigToken{}, errors.New("error getting token from rancher"))
			})

			It("returns an internal server error", func() {
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))
				b, _ := ioutil.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Error).To(Equal("error getting token from rancher"))
			})
		})

		When("token is expired", func() {
			BeforeEach(func() {
				// Call first time to get the expired token
				fakeRancherClient.NewTokenReturns(expiredRancherToken, nil)
				req, _ = http.NewRequest(http.MethodGet, uri, nil)
				res, err = http.DefaultClient.Do(req)
				// Call second time to get valid token
				fakeRancherClient.NewTokenReturns(validRancherToken, nil)
				req, _ = http.NewRequest(http.MethodGet, uri, nil)
				res, err = http.DefaultClient.Do(req)
				// Third call to get fake toke, but won't get called becase the cached token is valid
				fakeRancherClient.NewTokenReturns(fakeRancherToken, nil)
			})

			It("it succeeds", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				b, _ := ioutil.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Token).To(Equal("valid-rancher-token"))
			})
		})

		When("it succeeds", func() {
			It("succeeds", func() {
				Expect(res.StatusCode).To(Equal(http.StatusOK))
				b, _ := ioutil.ReadAll(res.Body)
				_ = json.Unmarshal(b, &tokens)
				Expect(tokens.Token).To(Equal("valid-rancher-token"))
			})
		})
	})
})
