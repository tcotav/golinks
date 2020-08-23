package routes

import (
	"errors"
	"net/url"
	"regexp"
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
	ShortKey       string
	URL            string
	Creator        string
	Team           string
	CreatedAt      time.Time
	ModifiedAt     time.Time
	LastModifiedBy string
}

var emailRegex = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")

// isEmailValid checks if the email provided passes the required structure and length.
func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}

// NewRoute is a
func NewRoute(k string, url string, creator string, team string) (Route, error) {
	now := time.Now()
	err := isValidURL(url)
	if err != nil {
		return Route{}, err
	}
	if !isEmailValid(creator) {
		return Route{}, errors.New("Invalid or bad format creator email address")
	}
	if team != "" && !isEmailValid(team) { // allow blank team
		return Route{}, errors.New("Invalid or bad format team email address")
	}
	return Route{ShortKey: k, URL: url, Creator: creator, Team: team,
		CreatedAt: now, ModifiedAt: now, LastModifiedBy: creator}, nil
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
