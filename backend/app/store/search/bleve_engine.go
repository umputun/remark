package search

import (
	"fmt"
	"github.com/blevesearch/bleve/v2"
	bleveCustom "github.com/blevesearch/bleve/v2/analysis/analyzer/custom"
	bleveStandard "github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	bleveEn "github.com/blevesearch/bleve/v2/analysis/lang/en"
	bleveRu "github.com/blevesearch/bleve/v2/analysis/lang/ru"
	"github.com/blevesearch/bleve/v2/analysis/token/lowercase"
	bleveSingle "github.com/blevesearch/bleve/v2/analysis/tokenizer/single"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/pkg/errors"
	"github.com/umputun/remark42/backend/app/store"
	"log"
	"os"
)

const userFieldName = "user"
const textFieldName = "text"

type bleveEngine struct {
	index     bleve.Index
	indexPath string
}

// Available text analyzers.
// Bleve supports a bit more languages that may be added,
// see https://github.com/blevesearch/bleve/tree/master/analysis/lang
var analyzerMapping = map[string]string{
	"standard": bleveStandard.Name,
	"english":  bleveEn.AnalyzerName,
	"russian":  bleveRu.AnalyzerName,
}

func newBleveEngine(indexPath, analyzer string) (eng *bleveEngine, isNew bool, err error) {
	if _, ok := analyzerMapping[analyzer]; !ok {
		analyzers := make([]string, 0, len(analyzerMapping))
		for k := range analyzerMapping {
			analyzers = append(analyzers, k)
		}
		return nil, isNew, errors.Errorf("Unknown analyzer: %q. Available analyzers for bleve: %v", analyzer, analyzers)
	}
	var index bleve.Index

	st, errOpen := os.Stat(indexPath)
	switch {
	case os.IsNotExist(errOpen):
		isNew = true
		// create new
		log.Printf("[INFO] creating new search index %s", indexPath)
		index, err = bleve.New(indexPath, createIndexMapping(analyzerMapping[analyzer]))
		if err != nil {
			return nil, isNew, errors.Wrap(err, "cannot open index")
		}
	case errOpen == nil:

		// open existing
		if !st.IsDir() {
			return nil, isNew, errors.Errorf("index path should be a directory")
		}
		log.Printf("[INFO] opening existing search index %s", indexPath)
		index, err = bleve.Open(indexPath)
		if err != nil {
			return nil, isNew, errors.Wrap(err, "cannot create index")
		}
	default:
		// error
		return nil, isNew, errors.Wrap(err, "cannot open index")
	}

	return &bleveEngine{
		index:     index,
		indexPath: indexPath,
	}, isNew, nil
}

// Index documents
func (b *bleveEngine) Index(comments []store.Comment) error {
	batch := b.index.NewBatch()
	for _, comment := range comments {
		err := batch.Index(comment.ID, comment)
		if err != nil {
			return errors.Wrap(err, "can't add to indexing batch")
		}
	}
	err := b.index.Batch(batch)
	if err != nil {
		return errors.Wrap(err, "index error")
	}
	return nil
}

// Search performs search request
func (b *bleveEngine) Search(req *Request) (*ResultPage, error) {
	bQuery := bleve.NewQueryStringQuery(req.Query)
	bReq := bleve.NewSearchRequestOptions(bQuery, req.Limit, req.From, false)

	if validateSortField(req.SortBy, "time") {
		bReq.SortBy([]string{req.SortBy})
	} else if req.SortBy != "" {
		log.Printf("[WARN] unknown sort field %q", req.SortBy)
	}

	bReq.Highlight = bleve.NewHighlight()
	bReq.Highlight.AddField(textFieldName)

	serp, err := b.index.Search(bReq)
	if err != nil {
		return nil, errors.Wrap(err, "bleve search error")
	}
	log.Printf("[INFO] found %d (limit %d) documents for query %q in %s",
		serp.Total, req.Limit, req.Query, serp.Took.String())

	result := convertBleveSerp(serp)
	return result, nil
}

// Delete document from index
func (b *bleveEngine) Delete(id string) error {
	return b.index.Delete(id)
}

// Close indexer
func (b *bleveEngine) Close() error {
	return b.index.Close()
}

// convertBleveSerp converts search result
// from bleve internal representation to general type ResultPage that would passed to user
func convertBleveSerp(bleveResult *bleve.SearchResult) *ResultPage {
	result := ResultPage{
		Total:     bleveResult.Total,
		Documents: make([]ResultDoc, 0, len(bleveResult.Hits)),
	}
	for _, r := range bleveResult.Hits {
		d := ResultDoc{
			ID:      r.ID,
			Matches: []TokenMatch{},
		}

		if highlight, has := r.Locations[textFieldName]; has {
			for _, locs := range highlight {
				for _, loc := range locs {
					d.Matches = append(d.Matches, TokenMatch{
						Start: loc.Start,
						End:   loc.End,
					})
				}
			}
		}

		result.Documents = append(result.Documents, d)
	}
	return &result
}

func validateSortField(sortBy string, possible ...string) bool {
	if sortBy == "" {
		return false
	}
	if sortBy[0] == '-' || sortBy[0] == '+' {
		sortBy = sortBy[1:]
	}
	for _, e := range possible {
		if sortBy == e {
			return true
		}
	}
	return false
}

func textMapping(analyzer string, doStore bool) *mapping.FieldMapping {
	textFieldMapping := bleve.NewTextFieldMapping()
	textFieldMapping.Store = doStore
	textFieldMapping.Analyzer = analyzer
	textFieldMapping.IncludeTermVectors = true
	return textFieldMapping
}

func createIndexMapping(textAnalyzer string) mapping.IndexMapping {
	indexMapping := bleve.NewIndexMapping()
	err := indexMapping.AddCustomAnalyzer("keyword_lower", map[string]interface{}{
		"type":      bleveCustom.Name,
		"tokenizer": bleveSingle.Name,
		"token_filters": []string{
			lowercase.Name,
		},
	})
	if err != nil {
		panic(fmt.Sprintf("error adding bleve analyzer %v", err))
	}

	// setup how comments would be indexed
	commentMapping := bleve.NewDocumentMapping()
	commentMapping.AddFieldMappingsAt(textFieldName, textMapping(textAnalyzer, false))

	userMapping := bleve.NewDocumentMapping()
	userMapping.AddFieldMappingsAt("name", textMapping("keyword_lower", true))
	commentMapping.AddSubDocumentMapping(userFieldName, userMapping)

	indexMapping.AddDocumentMapping("_default", commentMapping)

	return indexMapping
}
