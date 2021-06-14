package search

import (
	"encoding/hex"
	"github.com/pkg/errors"
	"github.com/umputun/remark42/backend/app/store"
	"hash/fnv"
	"path"
)

// Engine provides core functionality for searching
type Engine interface {
	Index(comments []*store.Comment) error
	Search(req *Request) (*ResultPage, error)
	Delete(id string) error
	Close() error
}

func newEngine(site string, params ServiceParams) (Engine, error) {
	// create separate folder for index for each site
	idxPath := path.Join(params.IndexPath, encodeSiteID(site))
	if params.Engine == "bleve" {
		return newBleveEngine(idxPath, params.Analyzer)
	}
	return nil, errors.Errorf("unknown search engine %q", params.Engine)
}

// encodeSiteID generate unique id for site that can be name of folder on fs
func encodeSiteID(siteID string) string {
	h := fnv.New32().Sum([]byte(siteID))
	return hex.EncodeToString(h)
}