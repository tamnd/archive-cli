package ia

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

const viewsTTL = time.Hour

// GetViews fetches short view statistics for one or more items in a single call.
func GetViews(ctx context.Context, h *HTTPClient, cache *Cache, identifiers []string) (map[string]Views, error) {
	key := strings.Join(identifiers, ",")
	u := ViewsURL + key
	var body []byte
	if cache != nil {
		if b, ok := cache.Get(u, viewsTTL); ok {
			body = b
		}
	}
	if body == nil {
		b, err := h.FetchBytes(ctx, u)
		if err != nil {
			return nil, err
		}
		body = b
		if cache != nil {
			cache.Put(u, body)
		}
	}
	out := map[string]Views{}
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	return out, nil
}
