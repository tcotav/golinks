package store

import (
	"database/sql"
	"errors"

	lru "github.com/hashicorp/golang-lru"
	"github.com/tcotav/golinks/routes"

	// database driver for sql package
	_ "github.com/mattn/go-sqlite3"
)

// See: https://golang.org/pkg/database/sql/

// Store is the data structure wrapping the underlying database interactions.  It contains
// the handle to the database, i.e. the database handle representing a pool of zero or
// more underlying connections.
type SQLStore struct {
	db    *sql.DB
	cache *lru.Cache
}

var (
	sharedStore *SQLStore = &SQLStore{}
)

// NewStore is the constructor for the Store struct and does the actual work of creating the DB handler.
func NewStore(dbConn *sql.DB) (*SQLStore, error) {
	if (SQLStore{}) != *sharedStore { // did we init store already
		return sharedStore, nil // if so, hand it back
	}
	cache, _ := lru.New(500)
	sharedStore = &SQLStore{db: dbConn, cache: cache}
	// end database setup
	return sharedStore, nil
}

const insertSQL = "INSERT INTO routes(short_key, url, creatorid, teamid, created_at, modified_at, last_modified_by) VALUES (?,?,?,?,?,?,?)"

func (s *SQLStore) Add(r routes.Route) (int, error) {
	stmt, err := s.db.Prepare(insertSQL)
	if err != nil {
		return -1, err
	}
	res, err := stmt.Exec(r.ShortKey, r.URL, r.CreatorID, r.TeamID, r.CreatedAt, r.ModifiedAt, r.LastModifiedByID)
	if err != nil {
		return -1, err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return int(affect), nil
}

const getRouteSQL = "SELECT short_key, url, creatorid, teamid, created_at, modified_at, last_modified_by FROM routes where short_key = ?"

// Get is
func (s *SQLStore) Get(k string) (routes.Route, error) {
	rows, err := s.db.Query(getRouteSQL, k)
	if err != nil {
		return routes.Route{}, err
	}
	defer rows.Close()

	var r routes.Route
	for rows.Next() {
		err := rows.Scan(&r.ShortKey, &r.URL, &r.CreatorID, &r.TeamID, &r.ModifiedAt, &r.LastModifiedByID)
		if err != nil {
			return routes.Route{}, err
		}
		return r, nil
	}
	return routes.Route{}, errors.New("No match found")
}

const getURLSQL = "SELECT  url FROM routes where short_key = ?"

func (s *SQLStore) GetURL(k string) (string, error) {
	// check local cache
	v, ok := s.cache.Get(k)
	if ok {
		return v.(string), nil
	}

	// then remote cache
	// then database
	rows, err := s.db.Query(getURLSQL, k)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var url string
	for rows.Next() {
		err := rows.Scan(&url)
		if err != nil {
			return "", err
		}
		// update local cache
		s.cache.Add(k, url)
		return url, nil
	}
	return "", errors.New("No match found")
}

func (s *SQLStore) Modify(routes.Route) error {
	return nil
}
func (s *SQLStore) Delete(string) bool {
	return false
}

/* round 2
func (s *RouteStore) GetAllForUser(string) []routes.Route{ }
func (s *RouteStore) GetRecentlyAdded() []routes.Route{}
func (s *RouteStore) GetRecentlyModified() []routes.Route {}
*/
