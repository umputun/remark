package api

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	bolt "github.com/coreos/bbolt"
	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/token"
	cache "github.com/go-pkgz/lcw"
	log "github.com/go-pkgz/lgr"
	R "github.com/go-pkgz/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/umputun/remark/backend/app/migrator"
	"github.com/umputun/remark/backend/app/notify"
	"github.com/umputun/remark/backend/app/rest"
	"github.com/umputun/remark/backend/app/rest/proxy"
	"github.com/umputun/remark/backend/app/store"
	adminstore "github.com/umputun/remark/backend/app/store/admin"
	"github.com/umputun/remark/backend/app/store/engine"
	"github.com/umputun/remark/backend/app/store/image"
	"github.com/umputun/remark/backend/app/store/service"
)

var devToken = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJyZW1hcms0MiIsImV4cCI6Mzc4OTE5MTgyMiwianRpIjoicmFuZG9tIGlkIiwiaXNzIjoicmVtYXJrNDIiLCJuYmYiOjE1MjE4ODQyMjIsInVzZXIiOnsibmFtZSI6ImRldmVsb3BlciBvbmUiLCJpZCI6ImRldiIsInBpY3R1cmUiOiJodHRwOi8vZXhhbXBsZS5jb20vcGljLnBuZyIsImlwIjoiMTI3LjAuMC4xIiwiZW1haWwiOiJtZUBleGFtcGxlLmNvbSJ9fQ.aKUAXiZxXypgV7m1wEOgUcyPOvUDXHDi3A06YWKbcLg`

var anonToken = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJyZW1hcms0MiIsImV4cCI6Mzc4OTE5MTgyMiwianRpIjoicmFuZG9tIGlkIiwiaXNzIjoicmVtYXJrNDIiLCJuYmYiOjE1MjE4ODQyMjIsInVzZXIiOnsibmFtZSI6ImFub255bW91cyB0ZXN0IHVzZXIiLCJpZCI6ImFub255bW91c190ZXN0X3VzZXIiLCJwaWN0dXJlIjoiaHR0cDovL2V4YW1wbGUuY29tL3BpYy5wbmciLCJpcCI6IjEyNy4wLjAuMSIsImVtYWlsIjoiYW5vbkBleGFtcGxlLmNvbSJ9fQ.gAae2WMxZNZE5ebVboptPEyQ7Nk6EQxciNnGJ_mPOuU`

var devTokenBadAud = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJyZW1hcms0Ml9iYWQiLCJleHAiOjM3ODkxOTE4MjIsImp0aSI6InJhbmRvbSBpZCIsImlzcyI6InJlbWFyazQyIiwibmJmIjoxNTIxODg0MjIyLCJ1c2VyIjp7Im5hbWUiOiJkZXZlbG9wZXIgb25lIiwiaWQiOiJkZXYiLCJwaWN0dXJlIjoiaHR0cDovL2V4YW1wbGUuY29tL3BpYy5wbmciLCJpcCI6IjEyNy4wLjAuMSIsImVtYWlsIjoibWVAZXhhbXBsZS5jb20ifX0.FuTTocVtcxr4VjpfIICvU2yOb3su28VkDzj94H9Q3xY`

var adminUmputunToken = `eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiJyZW1hcms0MiIsImV4cCI6MTk1NDU5Nzk4MCwianRpIjoiOTdhMmUwYWM0ZGM3ZDVmNjkyNmQ1ZTg2MjBhY2VmOWE0MGMwIiwiaWF0IjoxNDU0NTk3NjgwLCJpc3MiOiJyZW1hcms0MiIsInVzZXIiOnsibmFtZSI6IlVtcHV0dW4iLCJpZCI6ImdpdGh1Yl9lZjBmNzA2YTciLCJwaWN0dXJlIjoiaHR0cHM6Ly9yZW1hcms0Mi5yYWRpby10LmNvbS9hcGkvdjEvYXZhdGFyL2NiNDJmZjQ5M2FkZTY5NmQ4OGEzYTU5MGYxMzZhZTllMzRkZTdjMWIuaW1hZ2UiLCJhdHRycyI6eyJhZG1pbiI6dHJ1ZSwiYmxvY2tlZCI6ZmFsc2V9fX0.dZiOjWHguo9f42XCMooMcv4EmYFzifl_-LEvPZHCtks`

func TestRest_FileServer(t *testing.T) {
	ts, _, teardown := startupT(t)
	defer teardown()

	testHtmlName := "test-remark.html"
	testHTMLFile := os.TempDir() + "/" + testHtmlName
	err := ioutil.WriteFile(testHTMLFile, []byte("some html"), 0700)
	assert.NoError(t, err)

	body, code := get(t, ts.URL+"/web/"+testHtmlName)
	assert.Equal(t, 200, code)
	assert.Equal(t, "some html", body)
	_ = os.Remove(testHTMLFile)
}

func TestRest_GetStarted(t *testing.T) {
	ts, _, teardown := startupT(t)
	defer teardown()

	getStartedHTML := os.TempDir() + "/getstarted.html"
	err := ioutil.WriteFile(getStartedHTML, []byte("some html blah"), 0700)
	assert.NoError(t, err)

	body, code := get(t, ts.URL+"/index.html")
	assert.Equal(t, 200, code)
	assert.Equal(t, "some html blah", body)

	_ = os.Remove(getStartedHTML)
	_, code = get(t, ts.URL+"/index.html")
	assert.Equal(t, 404, code)

}

func TestRest_Shutdown(t *testing.T) {
	srv := Rest{Authenticator: &auth.Service{}, ImageProxy: &proxy.Image{}}
	finished := make(chan bool)

	// without waiting for channel close at the end goroutine will stay alive after test finish
	// which would create data race with next test
	go func() {
		time.Sleep(200 * time.Millisecond)
		srv.Shutdown()
		close(finished)
	}()

	st := time.Now()
	srv.Run(0)
	assert.True(t, time.Since(st).Seconds() < 1, "should take about 100ms")
	<-finished
}

func TestRest_filterComments(t *testing.T) {
	user := store.User{ID: "user1", Name: "user name 1"}
	c1 := store.Comment{User: user, Text: "test test #1", Locator: store.Locator{SiteID: "radio-t",
		URL: "https://radio-t.com/blah1"}, Timestamp: time.Date(2018, 05, 27, 1, 14, 10, 0, time.Local)}
	c2 := store.Comment{User: user, Text: "test test #2", ParentID: "p1", Locator: store.Locator{SiteID: "radio-t",
		URL: "https://radio-t.com/blah1"}, Timestamp: time.Date(2018, 05, 27, 1, 14, 20, 0, time.Local)}
	c3 := store.Comment{User: user, Text: "test test #3", ParentID: "p1", Locator: store.Locator{SiteID: "radio-t",
		URL: "https://radio-t.com/blah1"}, Timestamp: time.Date(2018, 05, 27, 1, 14, 25, 0, time.Local)}

	r := filterComments([]store.Comment{c1, c2, c3}, func(c store.Comment) bool {
		return c.Text == "test test #1" || c.Text == "test test #3"
	})
	assert.Equal(t, 2, len(r), "one comment filtered")
}

func TestRest_RunStaticSSLMode(t *testing.T) {
	sslPort := chooseRandomUnusedPort()
	srv := Rest{
		Authenticator: auth.NewService(auth.Opts{
			AvatarStore:       avatar.NewLocalFS("/tmp"),
			AvatarResizeLimit: 300,
		}),

		ImageProxy: &proxy.Image{},
		SSLConfig: SSLConfig{
			SSLMode: Static,
			Port:    sslPort,
			Key:     "../../cmd/testdata/key.pem",
			Cert:    "../../cmd/testdata/cert.pem",
		},
		RemarkURL: fmt.Sprintf("https://localhost:%d", sslPort),
	}

	port := chooseRandomUnusedPort()
	go func() {
		srv.Run(port)
	}()

	waitForHTTPSServerStart(sslPort)

	client := http.Client{
		// prevent http redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},

		// allow self-signed certificate
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/blah?param=1", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 307, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("https://localhost:%d/blah?param=1", sslPort), resp.Header.Get("Location"))

	resp, err = client.Get(fmt.Sprintf("https://localhost:%d/ping", sslPort))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "pong", string(body))

	srv.Shutdown()
}

func TestRest_RunAutocertModeHTTPOnly(t *testing.T) {
	sslPort := chooseRandomUnusedPort()
	srv := Rest{
		Authenticator: &auth.Service{},
		ImageProxy:    &proxy.Image{},
		SSLConfig: SSLConfig{
			SSLMode: Auto,
			Port:    sslPort,
		},
		RemarkURL: fmt.Sprintf("https://localhost:%d", sslPort),
	}

	port := chooseRandomUnusedPort()
	go func() {
		// can't check https server locally, just only http server
		srv.Run(port)
	}()

	waitForHTTPSServerStart(sslPort)

	client := http.Client{
		// prevent http redirect
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%d/blah?param=1", port))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 307, resp.StatusCode)
	assert.Equal(t, fmt.Sprintf("https://localhost:%d/blah?param=1", sslPort), resp.Header.Get("Location"))

	srv.Shutdown()
}

func TestRest_rejectAnonUser(t *testing.T) {

	ts := httptest.NewServer(fakeAuth(rejectAnonUser(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello")
	}))))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "use not logged in")

	resp, err = http.Get(ts.URL + "?fake_id=anonymous_user123&fake_name=test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "anon rejected")

	resp, err = http.Get(ts.URL + "?fake_id=real_user123&fake_name=test")
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "real user")
}

func Test_URLKey(t *testing.T) {
	tbl := []struct {
		url  string
		user store.User
		key  string
	}{
		{"http://example.com/1", store.User{}, "http://example.com/1"},
		{"http://example.com/1", store.User{ID: "user"}, "http://example.com/1"},
		{"http://example.com/1", store.User{ID: "user", Admin: true}, "admin!!http://example.com/1"},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r, err := http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)
			if tt.user.ID != "" {
				r = rest.SetUserInfo(r, tt.user)
			}
			assert.Equal(t, tt.key, URLKey(r))
		})
	}

}

func Test_URLKeyWithUser(t *testing.T) {
	tbl := []struct {
		url  string
		user store.User
		key  string
	}{
		{"http://example.com/1", store.User{}, "http://example.com/1"},
		{"http://example.com/1", store.User{ID: "user"}, "user!!http://example.com/1"},
		{"http://example.com/2", store.User{ID: "user2"}, "user2!!http://example.com/2"},
		{"http://example.com/1", store.User{ID: "user", Admin: true}, "admin!!user!!http://example.com/1"},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			r, err := http.NewRequest("GET", tt.url, nil)
			require.NoError(t, err)
			if tt.user.ID != "" {
				r = rest.SetUserInfo(r, tt.user)
			}
			assert.Equal(t, tt.key, URLKeyWithUser(r))
		})
	}

}

func TestRest_parseError(t *testing.T) {
	tbl := []struct {
		err error
		res int
	}{
		{errors.New("can not vote for his own comment"), rest.ErrVoteSelf},
		{errors.New("already voted for"), rest.ErrVoteDbl},
		{errors.New("maximum number of votes exceeded for comment"), rest.ErrVoteMax},
		{errors.New("minimal score reached for comment"), rest.ErrVoteMinScore},
		{errors.New("too late to edit"), rest.ErrCommentEditExpired},
		{errors.New("parent comment with reply can't be edited"), rest.ErrCommentEditChanged},
		{errors.New("blah blah"), rest.ErrInternal},
	}

	for n, tt := range tbl {
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			res := parseError(tt.err, rest.ErrInternal)
			assert.Equal(t, tt.res, res)
		})
	}
}

func startupT(t *testing.T) (ts *httptest.Server, srv *Rest, teardown func()) {
	log.Setup(log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)

	tmp := os.TempDir()
	var testDb string
	// pick a file name which is not in use for sure
	for i := 0; i < 10; i++ {
		testDb = fmt.Sprintf("/%s/test-remark-%d.db", tmp, rand.Int31())
		_, err := os.Stat(testDb)
		if err != nil {
			break
		}
	}
	_ = os.RemoveAll(tmp + "/ava-remark42")
	_ = os.RemoveAll(tmp + "/pics-remark42")

	b, err := engine.NewBoltDB(bolt.Options{}, engine.BoltSite{FileName: testDb, SiteID: "remark42"})
	require.NoError(t, err)

	cacheBackend, err := cache.NewExpirableCache()
	require.NoError(t, err)
	memCache := cache.NewScache(cacheBackend)

	astore := adminstore.NewStaticStore("123456", []string{"remark42"}, []string{"a1", "a2"}, "admin@remark-42.com")
	restrictedWordsMatcher := service.NewRestrictedWordsMatcher(service.StaticRestrictedWordsLister{Words: []string{"duck"}})

	dataStore := &service.DataStore{
		Engine:                 b,
		EditDuration:           5 * time.Minute,
		MaxCommentSize:         4000,
		AdminStore:             astore,
		MaxVotes:               service.UnlimitedVotes,
		RestrictedWordsMatcher: restrictedWordsMatcher,
	}

	srv = &Rest{
		DataService: dataStore,
		Authenticator: auth.NewService(auth.Opts{
			AdminPasswd:  "password",
			SecretReader: token.SecretFunc(func() (string, error) { return "secret", nil }),
			AvatarStore:  avatar.NewLocalFS(tmp + "/ava-remark42"),
		}),
		Cache:     memCache,
		WebRoot:   tmp,
		RemarkURL: "https://demo.remark42.com",
		ImageService: &image.Service{
			Store: &image.FileSystem{
				Location:   tmp + "/pics-remark42",
				Partitions: 100,
				MaxSize:    10000,
				Staging:    tmp + "/pics-remark42/staging",
			},
			TTL: time.Millisecond * 100,
		},
		ImageProxy:       &proxy.Image{},
		ReadOnlyAge:      10,
		CommentFormatter: store.NewCommentFormatter(&proxy.Image{}),
		Migrator: &Migrator{
			DisqusImporter:    &migrator.Disqus{DataStore: dataStore},
			WordPressImporter: &migrator.WordPress{DataStore: dataStore},
			NativeImporter:    &migrator.Native{DataStore: dataStore},
			NativeExporter:    &migrator.Native{DataStore: dataStore},
			UrlMapperMaker:    migrator.NewUrlMapper,
			Cache:             memCache,
			KeyStore:          astore,
		},
		Streamer: &Streamer{
			Refresh:   100 * time.Millisecond,
			TimeOut:   5 * time.Second,
			MaxActive: 100,
		},
		NotifyService: notify.NopService,
		EmojiEnabled:  true,
	}
	srv.ScoreThresholds.Low, srv.ScoreThresholds.Critical = -5, -10

	ts = httptest.NewServer(srv.routes())

	teardown = func() {
		ts.Close()
		require.NoError(t, srv.DataService.Close())
		_ = os.Remove(testDb)
		_ = os.RemoveAll(tmp + "/ava-remark42")
		_ = os.RemoveAll(tmp + "/pics-remark42")
	}

	return ts, srv, teardown
}

// fake auth middleware make user authenticated and uses query's fake_id for ID and fake_name for Name
func fakeAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("fake_id") != "" {
			r = rest.SetUserInfo(r, store.User{
				ID:   r.URL.Query().Get("fake_id"),
				Name: r.URL.Query().Get("fake_name"),
			})
		}
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func get(t *testing.T, url string) (string, int) {
	r, err := http.Get(url)
	require.NoError(t, err)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	require.NoError(t, err)
	return string(body), r.StatusCode
}

func sendReq(_ *testing.T, r *http.Request, token string) (*http.Response, error) {
	client := http.Client{Timeout: 5 * time.Second}
	if token != "" {
		r.Header.Set("X-JWT", token)
	}
	return client.Do(r)
}

func getWithDevAuth(t *testing.T, url string) (body string, code int) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	req.Header.Add("X-JWT", devToken)
	r, err := client.Do(req)
	require.NoError(t, err)
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	return string(b), r.StatusCode
}

func getWithAdminAuth(t *testing.T, url string) (string, int) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	req.SetBasicAuth("admin", "password")
	r, err := client.Do(req)
	require.NoError(t, err)
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	assert.NoError(t, err)
	return string(body), r.StatusCode
}
func post(t *testing.T, url string, body string) (*http.Response, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	assert.NoError(t, err)
	req.SetBasicAuth("admin", "password")
	return client.Do(req)
}

func addComment(t *testing.T, c store.Comment, ts *httptest.Server) string {
	b, err := json.Marshal(c)
	require.NoError(t, err, "can't marshal comment %+v", c)

	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", ts.URL+"/api/v1/comment", bytes.NewBuffer(b))
	require.NoError(t, err)
	req.Header.Add("X-JWT", devToken)
	resp, err := client.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	b, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	crResp := R.JSON{}
	err = json.Unmarshal(b, &crResp)
	require.NoError(t, err)
	time.Sleep(time.Nanosecond * 10)
	return crResp["id"].(string)
}

func requireAdminOnly(t *testing.T, req *http.Request) {
	resp, err := sendReq(t, req, "") // no-auth user
	require.NoError(t, err)
	assert.Equal(t, 401, resp.StatusCode)

	resp, err = sendReq(t, req, devToken) // non-admin user
	require.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func chooseRandomUnusedPort() (port int) {
	for i := 0; i < 10; i++ {
		port = 40000 + int(rand.Int31n(10000))
		if ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port)); err == nil {
			_ = ln.Close()
			break
		}
	}
	return port
}

func waitForHTTPSServerStart(port int) {
	// wait for up to 3 seconds for HTTPS server to start
	for i := 0; i < 300; i++ {
		time.Sleep(time.Millisecond * 10)
		conn, _ := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), time.Millisecond*10)
		if conn != nil {
			_ = conn.Close()
			break
		}
	}
}
