package routes

import (
	"encoding/json"
	"errors"
	"fmt"
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
	ID             int    `json:"id,omitempty"`
	ShortKey       string `json:"shortkey"`
	URL            string `json:"url"`
	Creator        string `json:"creator"`
	Team           string `json:"team,omitempty"`
	CreatedAt      string `json:"createdat,omitempty"`
	ModifiedAt     string `json:"modifiedat,omitempty"`
	LastModifiedBy string `json:"lastmodifiedby,omitempty"`
	Locked         int    `json:"locked"` // we will have some entries that will require elevated privs to change
}

const TimeFormat string = "2006-01-02 15:04:05"

// naive email regex
var emailRegex = regexp.MustCompile("^\\w+([\\.-]?\\w+)*@\\w+([\\.-]?\\w+)*(\\.\\w{2,3})+$")

// isEmailValid checks if the email provided passes the required structure and length.
func isEmailValid(e string) bool {
	if len(e) < 3 && len(e) > 254 {
		return false
	}
	return emailRegex.MatchString(e)
}

// NewRoute is a
func NewRoute(k string, url string, creator string, team string) (Route, error) {
	now := time.Now().Format(TimeFormat)

	err := isValidURL(url)
	if err != nil {
		return Route{}, err
	}
	if !isEmailValid(creator) {
		return Route{}, errors.New("Invalid or bad format creator email address")
	}
	if !isEmailValid(team) { // allow blank team
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

func (r *Route) UnmarshalJSON(body []byte) (err error) {
	// here we declare a local type, unmarshal data into it, and then convert back to desired type
	// in order to avoid a stack overflow with Decode/Unmarshal
	//
	// https://stackoverflow.com/questions/34859449/unmarshaljson-results-in-stack-overflow
	//
	type localr Route
	rLocal := localr{}
	if err := json.Unmarshal(body, &rLocal); err != nil {
		return err
	}

	route := Route(rLocal)
	r.ShortKey = route.ShortKey

	// expect valid url
	r.URL = route.URL
	if err != nil {
		return err
	}

	now := time.Now().Format(TimeFormat)

	// expect valid dates
	r.CreatedAt = route.CreatedAt
	if r.CreatedAt == "" {
		r.CreatedAt = now
	}
	r.ModifiedAt = route.ModifiedAt
	if r.ModifiedAt == "" {
		r.ModifiedAt = now
	}

	// expect valid email addresses
	r.Creator = route.Creator
	if !isEmailValid(r.Creator) {
		return fmt.Errorf("Invalid or bad format creator email address, %s", r.Creator)
	}

	r.LastModifiedBy = route.LastModifiedBy
	if r.LastModifiedBy == "" {
		r.LastModifiedBy = r.Creator
	} else if !isEmailValid(r.LastModifiedBy) {
		return fmt.Errorf("Invalid or bad format lastmodified by email address, %s", r.LastModifiedBy)
	}
	r.Team = route.Team
	if r.Team == "" {
		r.Team = r.Creator
	} else if !isEmailValid(r.Team) { // allow blank team
		return fmt.Errorf("Invalid or bad format team by email address, %s", r.Team)
	}
	return nil
}
