package search

import (
	"github.com/pkg/errors"
	"github.com/umputun/remark42/backend/app/store"
	"github.com/umputun/remark42/backend/app/store/engine"
)

// Request search query
type Request struct {
	SiteID string
	Query  string
	SortBy string
	From   int
	Limit  int
}

// TokenMatch describes match position
type TokenMatch struct {
	Start uint64 `json:"start"`
	End   uint64 `json:"end"`
}

// ResultDoc search result document
type ResultDoc struct {
	ID      string       `json:"id"`
	Matches []TokenMatch `json:"matches"`
}

// ResultPage returned from search
type ResultPage struct {
	Total     uint64      `json:"total"`
	Documents []ResultDoc `json:"documents"`
}

// IndexingStoreProxy is the wrapper and proxies all calls to engine.Interface and updates search index
type IndexingStoreProxy struct {
	engine.Interface
	search *Service
}

// WrapStoreEngineWithIndexer creates new engine.Interface that keep index conistent with updates
// and proxy all calls to underlying engine.Interface
func WrapStoreEngineWithIndexer(e engine.Interface, search *Service) engine.Interface {
	if search == nil {
		return e
	}
	return &IndexingStoreProxy{e, search}
}

// Create new comment and add it to search index
func (e *IndexingStoreProxy) Create(comment store.Comment) (commentID string, err error) {
	commentID, err = e.Interface.Create(comment)
	if err != nil {
		return "", err
	}

	err = e.search.Index(comment)
	if err != nil {
		return "", errors.Wrap(err, "failed to add document to search index")
	}
	return commentID, err
}

// Update comment and reindex it in search index
func (e *IndexingStoreProxy) Update(comment store.Comment) error {
	err := e.Interface.Update(comment)
	if err != nil {
		return err
	}
	err = e.search.Index(comment)
	if err != nil {
		return errors.Wrap(err, "failed to update document at search index")
	}
	return nil
}

// Delete comment from store and search index
func (e *IndexingStoreProxy) Delete(req engine.DeleteRequest) error {
	err := e.Interface.Delete(req)
	if err != nil {
		return err
	}
	err = e.search.Delete(req.Locator.SiteID, req.CommentID)
	if err != nil {
		return errors.Wrap(err, "failed to delete from search index")
	}
	return nil
}
