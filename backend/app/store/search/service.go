package search

import (
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/umputun/remark42/backend/app/store"
	"log"
)

// Service provides search functionality
type Service struct {
	shards map[string]Engine
}

// ServiceParams contains configuration for search service
type ServiceParams struct {
	Engine    string
	IndexPath string
	Analyzer  string
}

// NewService creates new search service
func NewService(sites []string, params ServiceParams) (*Service, error) {
	s := &Service{
		shards: map[string]Engine{},
	}
	for _, site := range sites {
		var err error
		s.shards[site], err = newEngine(site, params)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create search engine")
		}
	}
	return s, nil
}

// Search document
func (s *Service) Search(req *Request) (*ResultPage, error) {
	if eng, has := s.shards[req.SiteID]; has {
		return eng.Search(req)
	}
	return nil, errors.Errorf("no search index for site %q", req.SiteID)
}

// Index single document
func (s *Service) Index(doc *store.Comment) error {
	return s.IndexBatch([]*store.Comment{doc})
}

// IndexBatch indexes batch of document
func (s *Service) IndexBatch(docs []*store.Comment) error {
	if len(docs) == 0 {
		return nil
	}
	siteID := docs[0].Locator.SiteID
	if eng, has := s.shards[siteID]; has {
		// validate that all documents from same site
		for _, doc := range docs {
			if doc.Locator.SiteID != siteID {
				return errors.Errorf("different sites in batch")
			}
		}
		return eng.Index(docs)
	}
	return errors.Errorf("site %q not found", siteID)
}

// Delete document from index
func (s *Service) Delete(siteID, commentID string) error {
	if eng, has := s.shards[siteID]; has {
		return eng.Delete(commentID)
	}
	return errors.Errorf("Site %q not found", siteID)
}

// Close search service
func (s *Service) Close() error {

	log.Print("[INFO] closing search service...")
	errs := new(multierror.Error)

	for siteID, searcher := range s.shards {
		if err := searcher.Close(); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "cannot close searcher for %q", siteID))
		}
	}
	log.Print("[INFO] search service closed")
	return errs.ErrorOrNil()
}
