package routes

import (
	"net/url"
	"time"
)

/*
#### Data structure:
- short_key - key used in url for lookup
- url - returned url correlated to the key
- creator - creator of the k+v
- created_at - date of creation
- modified_at - date last modified
- last_modified_by - user last modified key
*/

// Route is
type Route struct {
	ShortKey         string
	URL              string
	CreatorID        int
	TeamID           int
	CreatedAt        time.Time
	ModifiedAt       time.Time
	LastModifiedByID int
}

// NewRoute is a
func NewRoute(k string, url string, creatorID int, teamID int) (Route, error) {
	now := time.Now()
	err := isValidURL(url)
	if err != nil {
		return Route{}, err
	}

	return Route{ShortKey: k, URL: url, CreatorID: creatorID, TeamID: teamID,
		CreatedAt: now, ModifiedAt: now, LastModifiedByID: creatorID}, nil
}

// isValidURL is
func isValidURL(testURL string) error {
	_, err := url.ParseRequestURI(testURL)
	if err != nil {
		return err
	}

	u, err := url.Parse(testURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return err
	}

	return nil
}
