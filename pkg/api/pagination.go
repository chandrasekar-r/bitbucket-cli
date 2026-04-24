package api

import (
	"encoding/json"
	"fmt"
)

// Page represents a single Bitbucket paginated response envelope.
// Bitbucket uses offset-based pagination with a max of 100 items per page.
//
// All list endpoints return this shape:
//
//	{"pagelen":10, "page":1, "size":47, "next":"...", "values":[...]}
type Page struct {
	Pagelen int             `json:"pagelen"`
	Page    int             `json:"page"`
	Size    int             `json:"size"`
	Next    string          `json:"next"`
	Values  json.RawMessage `json:"values"`
}

// PaginateAll fetches all pages from startURL, appending each page's values
// into a single JSON array. Stops after collecting limit items (0 = no limit).
//
// The returned []byte is a valid JSON array, e.g. [{"slug":"repo1"}, ...].
func PaginateAll(client *Client, startURL string, limit int) ([]json.RawMessage, error) {
	var all []json.RawMessage
	url := startURL

	for url != "" {
		// Strip base URL prefix if the `next` field returns an absolute URL
		path := stripBase(client.baseURL, url)

		var page Page
		if err := client.Get(path, &page); err != nil {
			return nil, fmt.Errorf("fetching page: %w", err)
		}

		var items []json.RawMessage
		if err := json.Unmarshal(page.Values, &items); err != nil {
			return nil, fmt.Errorf("parsing page values: %w", err)
		}

		for _, item := range items {
			all = append(all, item)
			if limit > 0 && len(all) >= limit {
				return all, nil
			}
		}

		url = page.Next
	}
	return all, nil
}

// stripBase removes the baseURL prefix from an absolute URL returned by the
// Bitbucket API's `next` field, leaving a path-only string for Client.Get.
func stripBase(baseURL, url string) string {
	if len(url) > len(baseURL) && url[:len(baseURL)] == baseURL {
		return url[len(baseURL):]
	}
	return url
}
