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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.homedepot.com/cd/arcade/pkg/google/googlefakes"
	arcade_http "github.homedepot.com/cd/arcade/pkg/http"
	"github.homedepot.com/cd/arcade/pkg/middleware"
	"github.homedepot.com/cd/arcade/pkg/rancher"
	"github.homedepot.com/cd/arcade/pkg/rancher/rancherfakes"
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
	fakeGoogleToken   = "fake-google-token"
	fakeRancherClient *rancherfakes.FakeClient
	fakeRancherToken  = rancher.KubeconfigToken{
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
	Describe("#GetToken", func() {
		When("provider is not supported", func() {
			BeforeEach(func() {
				r := gin.New()
				r.Use(gin.Recovery())
				r.GET("/tokens", arcade_http.GetToken)

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
			r.GET("/tokens", arcade_http.GetToken)

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
				fakeGoogleClient.NewTokenReturns("", errors.New("error getting token from google"))
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
			r.GET("/tokens", arcade_http.GetToken)

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
				r.GET("/tokens", arcade_http.GetToken)

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
