package search

import (
	"context"
	"github.com/pkg/errors"
	"github.com/umputun/remark42/backend/app/store"
	"github.com/umputun/remark42/backend/app/store/engine"
	"log"
	"time"
)

// StartupIndex performs indexing at first with existing data
func StartupIndex(ctx context.Context, siteID string, s *Service, e engine.Interface) error {
	if _, isIndexed := s.indexedSites[siteID]; isIndexed {
		log.Printf("[INFO] skipping indexing site %q", siteID)
		return nil
	}
	log.Printf("[INFO] indexing site %q", siteID)
	startTime := time.Now()

	req := engine.InfoRequest{Locator: store.Locator{SiteID: siteID}}
	topics, err := e.Info(req)

	if err != nil {
		return errors.Wrapf(err, "failed to get topics for site %q", siteID)
	}

	indexedCnt := 0
	for i := len(topics) - 1; i >= 0; i-- {
		locator := store.Locator{SiteID: siteID, URL: topics[i].URL}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req := engine.FindRequest{Locator: locator, Since: time.Time{}}
		comments, findErr := e.Find(req)
		if findErr != nil {
			return errors.Wrapf(findErr, "failed to fetch comments")
		}

		indexErr := s.IndexBatch(comments)
		log.Printf("[INFO] %d documents indexed site %v", len(comments), locator)

		if indexErr != nil {
			return errors.Wrapf(indexErr, "failed to index comments")
		}
		indexedCnt += len(comments)
	}
	log.Printf("[INFO] total %d documents indexed for site %q in %v", indexedCnt, siteID, time.Since(startTime))
	return err
}
