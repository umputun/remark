package search

import (
	"fmt"
	log "github.com/go-pkgz/lgr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/umputun/remark42/backend/app/store"
	"os"
	"testing"
	"time"
)

func createTestService(t *testing.T, sites []string) (searcher *Service, teardown func()) {
	tmp := os.TempDir()
	idxPath := tmp + "/search-remark42"

	_ = os.RemoveAll(idxPath)

	searcher, err := NewService(sites, ServiceParams{
		Engine:      "bleve",
		IndexPath: idxPath,
		Analyzer:  "standard",
	})

	require.NoError(t, err)

	teardown = func() {
		err := searcher.Close()
		require.NoError(t, err)
		_ = os.RemoveAll(idxPath)
	}
	return searcher, teardown
}

func TestSearch_SiteMux(t *testing.T) {
	searcher, teardown := createTestService(t, []string{"test-site", "test-site2", "test-site3"})
	defer teardown()

	err := searcher.Index(&store.Comment{
		ID:        "123456",
		Locator:   store.Locator{SiteID: "test-site", URL: "http://example.com/post1"},
		Text:      "text 123",
		User:      store.User{ID: "u1", Name: "user1"},
		Timestamp: time.Date(2017, 12, 20, 15, 18, 24, 0, time.Local),
	})
	assert.NoError(t, err)
	err = searcher.Index(&store.Comment{
		ID:        "123456",
		Locator:   store.Locator{SiteID: "test-site2", URL: "http://example.com/post1"},
		Text:      "text 345",
		User:      store.User{ID: "u1", Name: "user1"},
		Timestamp: time.Date(2017, 12, 20, 15, 20, 24, 0, time.Local),
	})
	assert.NoError(t, err)
	err = searcher.Index(&store.Comment{
		ID:        "123457",
		Locator:   store.Locator{SiteID: "test-site2", URL: "http://example.com/post1"},
		Text:      "foobar 345",
		User:      store.User{ID: "u1", Name: "user1"},
		Timestamp: time.Date(2017, 12, 20, 15, 20, 28, 0, time.Local),
	})
	assert.NoError(t, err)

	{
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "123", Limit: 3})
		require.NoError(t, err)
		require.Len(t, res.Documents, 1)
		assert.Equal(t, "123456", res.Documents[0].ID)

		require.Len(t, res.Documents[0].Matches, 1)
		assert.Equal(t, res.Documents[0].Matches[0], TokenMatch{Start: 5, End: 8})

		res, err = searcher.Search(&Request{SiteID: "test-site", Query: "345", Limit: 3})
		require.NoError(t, err)
		require.Len(t, res.Documents, 0)
	}
	{
		res, err := searcher.Search(&Request{SiteID: "test-site2", Query: "345", SortBy: "-time", Limit: 3})
		require.NoError(t, err)
		require.Len(t, res.Documents, 2)
		assert.Equal(t, "123457", res.Documents[0].ID)
		require.Len(t, res.Documents[0].Matches, 1)
		assert.Equal(t, res.Documents[0].Matches[0], TokenMatch{Start: 7, End: 10})

		assert.Equal(t, "123456", res.Documents[1].ID)
		require.Len(t, res.Documents[1].Matches, 1)
		assert.Equal(t, res.Documents[1].Matches[0], TokenMatch{Start: 5, End: 8})

		res, err = searcher.Search(&Request{SiteID: "test-site2", Query: "123", Limit: 3})
		require.NoError(t, err)
		assert.Len(t, res.Documents, 0)
	}
	{
		res, err := searcher.Search(&Request{SiteID: "test-site3", Query: "345", Limit: 3})
		require.NoError(t, err)
		assert.Len(t, res.Documents, 0)
	}
}

func TestSearch_Paginate(t *testing.T) {
	searcher, teardown := createTestService(t, []string{"test-site"})
	defer teardown()

	t0 := time.Date(2017, 12, 20, 15, 18, 24, 0, time.Local)
	for shift := 0; shift < 4; shift++ {
		cid := fmt.Sprintf("comment%d", shift)
		err := searcher.Index(&store.Comment{
			ID:        cid,
			Locator:   store.Locator{SiteID: "test-site", URL: fmt.Sprintf("http://example.com/post%d", shift%2)},
			Text:      "text 123",
			User:      store.User{ID: "u1", Name: "user1"},
			Timestamp: t0.Add(time.Duration(shift) * time.Minute),
		})
		assert.NoError(t, err)
	}

	{
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "123", Limit: 1, SortBy: "+time", From: 0})
		require.NoError(t, err)
		require.Len(t, res.Documents, 1)
		assert.Equal(t, "comment0", res.Documents[0].ID)
	}
	{
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "123", Limit: 1, SortBy: "+time", From: 1})
		require.NoError(t, err)
		require.Len(t, res.Documents, 1)
		assert.Equal(t, "comment1", res.Documents[0].ID)
	}
	{
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "123", Limit: 1, SortBy: "+time", From: 3})
		require.NoError(t, err)
		require.Len(t, res.Documents, 1)
		assert.Equal(t, "comment3", res.Documents[0].ID)
	}
	{
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "123", Limit: 2, From: 1, SortBy: "-time"})
		require.NoError(t, err)
		require.Len(t, res.Documents, 2)
		assert.Equal(t, []string{"comment2", "comment1"}, []string{res.Documents[0].ID, res.Documents[1].ID})
	}
}

func TestSearch_Delete(t *testing.T) {
	searcher, teardown := createTestService(t, []string{"test-site"})
	defer teardown()

	timestamp := time.Date(2017, 12, 20, 15, 18, 24, 0, time.Local)

	err := searcher.Index(&store.Comment{
		ID:        "comment1",
		Locator:   store.Locator{SiteID: "test-site", URL: "http://example.com/post"},
		Text:      "text 123",
		User:      store.User{ID: "u1", Name: "user1"},
		Timestamp: timestamp,
	})
	require.NoError(t, err)

	err = searcher.Index(&store.Comment{
		ID:        "comment2",
		Locator:   store.Locator{SiteID: "test-site", URL: "http://example.com/post"},
		Text:      "text 345",
		User:      store.User{ID: "u1", Name: "user1"},
		Timestamp: timestamp.Add(time.Hour),
	})
	require.NoError(t, err)

	{
		res, searchErr := searcher.Search(&Request{SiteID: "test-site", Query: "text", SortBy: "+time", Limit: 10})
		require.NoError(t, searchErr)
		require.Len(t, res.Documents, 2)
		assert.Equal(t, "comment1", res.Documents[0].ID)
		assert.Equal(t, "comment2", res.Documents[1].ID)
	}

	err = searcher.Delete("test-site", "comment1")
	require.NoError(t, err)

	{
		res, searchErr := searcher.Search(&Request{SiteID: "test-site", Query: "text", SortBy: "+time", Limit: 10})
		require.NoError(t, searchErr)
		require.Len(t, res.Documents, 1)
		assert.Equal(t, "comment2", res.Documents[0].ID)
	}
}

func TestSearch_OtherFields(t *testing.T) {
	searcher, teardown := createTestService(t, []string{"test-site", "test-site2", "test-site3"})
	defer teardown()

	err := searcher.Index(&store.Comment{
		ID:        "123456",
		Locator:   store.Locator{SiteID: "test-site", URL: "http://example.com/post1"},
		Text:      "text 123",
		User:      store.User{ID: "u1", Name: "user foo"},
		Timestamp: time.Date(2017, 12, 18, 15, 18, 24, 0, time.Local),
	})
	assert.NoError(t, err)
	err = searcher.Index(&store.Comment{
		ID:        "123457",
		Locator:   store.Locator{SiteID: "test-site", URL: "http://example.com/post1"},
		Text:      "text 345",
		User:      store.User{ID: "u2", Name: "User Bar"},
		Timestamp: time.Date(2017, 12, 21, 15, 20, 24, 0, time.Local),
	})
	assert.NoError(t, err)
	err = searcher.Index(&store.Comment{
		ID:        "123458",
		Locator:   store.Locator{SiteID: "test-site", URL: "http://example.com/post1"},
		Text:      "foobar text",
		User:      store.User{ID: "u2", Name: "User Bar"},
		Timestamp: time.Date(2017, 12, 25, 16, 20, 28, 0, time.Local),
	})
	assert.NoError(t, err)

	// username
	{
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "text +user.name:\"user bar\"", Limit: 20})
		require.NoError(t, err)
		assert.Len(t, res.Documents, 2)
	}
	{
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "text +user.name:\"user foo\"", Limit: 20})
		require.NoError(t, err)
		assert.Len(t, res.Documents, 1)
	}
	{
		// order matters in username field, match only whole token
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "text +user:\"foo user\"", Limit: 20})
		require.NoError(t, err)
		assert.Len(t, res.Documents, 0)
	}

	// time range
	{
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "text +time:>\"2017-12-20\"", Limit: 20})
		require.NoError(t, err)
		assert.Len(t, res.Documents, 2)
	}
	{
		res, err := searcher.Search(&Request{SiteID: "test-site", Query: "text +time:<\"2017-12-20\"", Limit: 20})
		require.NoError(t, err)
		assert.Len(t, res.Documents, 1)
	}
}

func TestMain(m *testing.M) {
	log.Setup(log.Debug, log.CallerFile, log.CallerFunc, log.Msec, log.LevelBraces)
	os.Exit(m.Run())
}

