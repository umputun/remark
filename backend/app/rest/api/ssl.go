package api

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"golang.org/x/crypto/acme/autocert"
)

// sslMode defines ssl mode for rest server
type sslMode int8

const (
	// None defines to run http server only
	None sslMode = iota

	// Static defines to run both https and http server. Redirect http to https
	Static

	// Auto defines to run both https and http server. Redirect http to https. Https server with autocert support
	Auto
)

// SSLConfig holds all ssl params for rest server
type SSLConfig struct {
	SSLMode      sslMode
	Cert         string
	Key          string
	Port         int
	ACMELocation string
}

// httpToHTTPSRouter creates new router which does redirect from http to https server
// with default middlewares. Used in 'static' ssl mode.
func (s *Rest) httpToHTTPSRouter() chi.Router {
	log.Printf("[DEBUG] create https-to-http redirect routes")
	router := chi.NewRouter()
	router.Use(middleware.RealIP, Recoverer)
	router.Use(middleware.Throttle(1000), middleware.Timeout(60*time.Second))

	router.Handle("/*", s.redirectHandler())
	return router
}

// httpChallengeRouter creates new router which perform ACME "http-01" challenge response
// with default middlewares. This part is necessary to obtain certificate from LE.
// If it receives not a acme challenge it performs redirect to https server.
// Used in 'auto' ssl mode.
func (s *Rest) httpChallengeRouter(m *autocert.Manager) chi.Router {
	log.Printf("[DEBUG] create http-challenge routes")
	router := chi.NewRouter()
	router.Use(middleware.RealIP, Recoverer)
	router.Use(middleware.Throttle(1000), middleware.Timeout(60*time.Second))

	router.Handle("/*", m.HTTPHandler(s.redirectHandler()))
	return router
}

func (s *Rest) redirectHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newURL := s.RemarkURL + r.URL.Path
		if r.URL.RawQuery != "" {
			newURL += "?" + r.URL.RawQuery
		}
		http.Redirect(w, r, newURL, http.StatusTemporaryRedirect)
	})
}

func (s *Rest) makeAutocertManager() *autocert.Manager {
	return &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		Cache:      autocert.DirCache(s.SSLConfig.ACMELocation),
		HostPolicy: autocert.HostWhitelist(s.getRemarkHost()),
		Email:      s.DataService.AdminStore.Email(""),
	}
}

// makeHTTPSAutoCertServer makes https server with autocert mode (LE support)
func (s *Rest) makeHTTPSAutocertServer(port int, router chi.Router, m *autocert.Manager) *http.Server {
	server := s.makeHTTPServer(port, router)
	server.TLSConfig = &tls.Config{GetCertificate: m.GetCertificate}
	return server
}

// getRemarkHost returns hostname for remark server.
// For example for remarkURL https://remark.com:443 it should return remark.com
func (s *Rest) getRemarkHost() string {
	u, err := url.Parse(s.RemarkURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}
