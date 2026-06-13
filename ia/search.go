package ia

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const searchTTL = 10 * time.Minute

// SearchQuery describes a search over the item store.
type SearchQuery struct {
	Query  string   // Lucene query
	Fields []string // fields to return (fl[])
	Sorts  []string // sort keys (sort[]), e.g. "downloads desc"
	Rows   int      // page size for Advanced Search
	Page   int      // 1-based page for Advanced Search
}

// DefaultFields are returned when the caller asks for none.
var DefaultFields = []string{"identifier", "title", "mediatype", "downloads", "date"}

// AdvancedSearchResult is the decoded Advanced Search response.
type AdvancedSearchResult struct {
	NumFound int
	Start    int
	Docs     []SearchDoc
}

// Search runs one page of Advanced Search.
func Search(ctx context.Context, h *HTTPClient, cache *Cache, q SearchQuery) (AdvancedSearchResult, error) {
	fields := q.Fields
	if len(fields) == 0 {
		fields = DefaultFields
	}
	rows := q.Rows
	if rows <= 0 {
		rows = 50
	}
	page := q.Page
	if page <= 0 {
		page = 1
	}
	v := url.Values{}
	v.Set("q", q.Query)
	for _, f := range fields {
		v.Add("fl[]", f)
	}
	for _, s := range q.Sorts {
		v.Add("sort[]", s)
	}
	v.Set("rows", strconv.Itoa(rows))
	v.Set("page", strconv.Itoa(page))
	v.Set("output", "json")
	u := AdvancedSearch + "?" + v.Encode()

	var body []byte
	if cache != nil {
		if b, ok := cache.Get(u, searchTTL); ok {
			body = b
		}
	}
	if body == nil {
		b, err := h.FetchBytes(ctx, u)
		if err != nil {
			return AdvancedSearchResult{}, err
		}
		body = b
		if cache != nil {
			cache.Put(u, body)
		}
	}
	var raw struct {
		Response struct {
			NumFound int         `json:"numFound"`
			Start    int         `json:"start"`
			Docs     []SearchDoc `json:"docs"`
		} `json:"response"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return AdvancedSearchResult{}, fmt.Errorf("decoding search response: %w", err)
	}
	if raw.Error != "" {
		return AdvancedSearchResult{}, fmt.Errorf("search error: %s", raw.Error)
	}
	return AdvancedSearchResult{NumFound: raw.Response.NumFound, Start: raw.Response.Start, Docs: raw.Response.Docs}, nil
}

// SearchEach pages through Advanced Search calling fn for each hit, up to limit
// (0 = the Solr deep-paging ceiling). It is the bounded, sortable path.
func SearchEach(ctx context.Context, h *HTTPClient, cache *Cache, q SearchQuery, limit int, fn func(SearchDoc) error) (int, error) {
	if q.Rows <= 0 {
		q.Rows = 100
		if limit > 0 && limit < 100 {
			q.Rows = limit
		}
	}
	q.Page = 1
	emitted := 0
	for {
		res, err := Search(ctx, h, cache, q)
		if err != nil {
			return emitted, err
		}
		if len(res.Docs) == 0 {
			return emitted, nil
		}
		for _, d := range res.Docs {
			if err := fn(d); err != nil {
				return emitted, err
			}
			emitted++
			if limit > 0 && emitted >= limit {
				return emitted, nil
			}
		}
		if res.Start+len(res.Docs) >= res.NumFound {
			return emitted, nil
		}
		q.Page++
	}
}

// Scrape exports every matching item via the cursor-based Scraping API, calling
// fn for each hit, up to limit (0 = all). This is the unbounded path used by
// --all; it does not support sort.
func Scrape(ctx context.Context, h *HTTPClient, q SearchQuery, limit int, fn func(SearchDoc) error) (int, error) {
	fields := q.Fields
	if len(fields) == 0 {
		fields = DefaultFields
	}
	count := 10000
	if limit > 0 && limit < count {
		count = max(limit, 100)
	}
	emitted := 0
	cursor := ""
	for {
		v := url.Values{}
		v.Set("q", q.Query)
		v.Set("fields", strings.Join(fields, ","))
		v.Set("count", strconv.Itoa(count))
		if cursor != "" {
			v.Set("cursor", cursor)
		}
		u := ScrapeURL + "?" + v.Encode()
		var page struct {
			Items  []SearchDoc `json:"items"`
			Count  int         `json:"count"`
			Cursor string      `json:"cursor"`
			Total  int         `json:"total"`
			Error  string      `json:"error"`
		}
		if err := h.GetJSON(ctx, u, &page); err != nil {
			return emitted, err
		}
		if page.Error != "" {
			return emitted, fmt.Errorf("scrape error: %s", page.Error)
		}
		for _, d := range page.Items {
			if err := fn(d); err != nil {
				return emitted, err
			}
			emitted++
			if limit > 0 && emitted >= limit {
				return emitted, nil
			}
		}
		if page.Cursor == "" || len(page.Items) == 0 {
			return emitted, nil
		}
		cursor = page.Cursor
	}
}
