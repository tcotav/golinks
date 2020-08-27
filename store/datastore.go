package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	lru "github.com/hashicorp/golang-lru"
	"github.com/tcotav/golinks/routes"

	// database driver for sql package

	"github.com/mattn/go-sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

// See: https://golang.org/pkg/database/sql/

// Store is the data structure wrapping the underlying database interactions.  It contains
// the handle to the database, i.e. the database handle representing a pool of zero or
// more underlying connections.
type DataStore struct {
	dbtype   string
	redisTTL time.Duration
	db       *sql.DB
	cache    *lru.Cache
	redis    *redis.Client
}

var (
	sharedStore *DataStore = &DataStore{}
)

// GetStore is the constructor for the Store struct and does the actual work of creating the DB handler.
func GetStore(dbtype string, dbConn *sql.DB, redisClient *redis.Client, redisTTL int) (*DataStore, error) {
	if (DataStore{}) != *sharedStore { // did we init store already
		return sharedStore, nil // if so, hand it back
	}

	return NewStore(dbtype, dbConn, redisClient, redisTTL)
}

func NewStore(dbtype string, dbConn *sql.DB, redisClient *redis.Client, redisTTL int) (*DataStore, error) {
	var cache *lru.Cache
	if redisClient == nil {
		cache, _ = lru.New(500)
	}
	newStore := &DataStore{dbtype: dbtype, db: dbConn, cache: cache, redis: redisClient, redisTTL: (time.Duration(redisTTL) * time.Second)}
	// end database setup
	return newStore, nil
}

const insertRoute = "INSERT INTO routes(short_key, url, creatorid, teamid, created_at, modified_at, last_modified_by) VALUES (?,?,?,?,?,?,?)"
const insertUser = "INSERT INTO users(name, created_at) VALUES(?,?)"
const getUser = "SELECT id FROM users where name = ?"

func (s *DataStore) GetUserID(username string) (int, error) {
	// create or get
	rows, err := s.db.Query(getUser, username)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	var userID int
	for rows.Next() {
		err := rows.Scan(&userID)
		if err != nil {
			return -1, err
		}
		return userID, nil
	}

	now := time.Now()
	stmt, err := s.db.Prepare(insertUser)
	if err != nil {
		return -1, err
	}
	res, err := stmt.Exec(username, now)
	if err != nil {
		return -1, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}
	return int(id), nil

}
func (s *DataStore) Add(r routes.Route) (int, error) {
	stmt, err := s.db.Prepare(insertRoute)
	if err != nil {
		return -1, err
	}
	res, err := stmt.Exec(r.ShortKey, r.URL, r.Creator, r.Team, r.CreatedAt, r.ModifiedAt, r.LastModifiedBy)
	if err != nil {
		return -1, err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return int(affect), nil
}

func (s *DataStore) GetRandomURL(k string) (routes.Route, error) {
	// get count of rows in url list
	//
	return routes.Route{}, nil
}

const getRouteSQL = "SELECT short_key, url, creatorid, teamid, created_at, modified_at, last_modified_by FROM routes where short_key = ?"

// Get is
func (s *DataStore) Get(k string) (routes.Route, error) {
	rows, err := s.db.Query(getRouteSQL, k)
	if err != nil {
		return routes.Route{}, err
	}
	defer rows.Close()

	var r routes.Route
	for rows.Next() {
		err := rows.Scan(&r.ShortKey, &r.URL, &r.Creator, &r.Team, &r.CreatedAt, &r.ModifiedAt, &r.LastModifiedBy)
		if err != nil {
			return routes.Route{}, err
		}
		return r, nil
	}
	return routes.Route{}, errors.New("No match found")
}

const getURLSQL = "SELECT  url FROM routes where short_key = ?"

func (s *DataStore) GetURL(k string) (string, error) {
	// check local cache
	if s.cache != nil {
		v, ok := s.cache.Get(k)
		if ok {
			return v.(string), nil
		}
	} else if s.redis != nil {
		sRet, _ := s.redis.Get(k).Result()
		if sRet != "" {
			// update lcl cache and return
			return sRet, nil
		}
	}

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

		if s.cache != nil {
			// update local cache
			s.cache.Add(k, url)
		} else if s.redis != nil {
			// update redis cache
			s.redis.Set(k, url, s.redisTTL).Err()
		}
		return url, nil
	}
	return "", errors.New("No match found")
}

const updateURLSQL = "UPDATE routes SET url=?, last_modified_by=?, modified_at=? where short_key = ?"

func (s *DataStore) Modify(r routes.Route) (int, error) {
	// what about case where we are changing the shortkey -- how to invalidate caches?
	now := time.Now()
	userID, err := s.GetUserID(r.LastModifiedBy)
	if err != nil {
		return -1, err
	}
	res, err := s.db.Exec(updateURLSQL, r.URL, now, userID, r.ShortKey)
	if err != nil {
		return -1, err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	// update the cache
	// is it in the cache
	// if it is, change the url
	if s.cache != nil {
		if _, ok := s.cache.Get(r.ShortKey); !ok {
			s.cache.Remove(r.ShortKey)
			s.cache.Add(r.ShortKey, r.URL)
		}
	} else if s.redis != nil {
		s.redis.Set(r.ShortKey, r.URL, s.redisTTL).Err()
	}
	return int(affect), nil
}

const deleteRouteSQL = "DELETE FROM routes where short_key = ?"

func (s *DataStore) Delete(k string) error {
	// what about case where we are changing the shortkey -- how to invalidate caches?
	res, err := s.db.Exec(deleteRouteSQL, k)
	if err != nil {
		return err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affect != 1 {
		return fmt.Errorf("Invalid delete for %s -- impacted %d rows", k, affect)
	}

	// update the cache
	// is it in the cache
	// if it is, change the url
	if s.cache != nil {
		if _, ok := s.cache.Get(k); !ok {
			s.cache.Remove(k)
		}
	} else if s.redis != nil {
		s.redis.Del(k).Err()
	}

	return nil
}

func (s *DataStore) IsSQLErrUniqueContraint(err error) bool {
	if s.dbtype == "sqlite" {
		if driverErr, ok := err.(sqlite3.Error); ok {
			if driverErr.ExtendedCode == sqlite3.ErrConstraintUnique {
				return true
			}
		}
		return false
	} else if s.dbtype == "mysql" {
		// 1062 duplicate key
		//
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == 1062 {
				return true
			}
		}
		return false
	}
	return false
}

func (s *DataStore) IsSQLErrDuplicateContraint(err error) bool {

	if s.dbtype == "sqlite" {
		// (1555) SQLITE_CONSTRAINT_PRIMARYKEY
		// (2579) SQLITE_CONSTRAINT_ROWID
		if driverErr, ok := err.(sqlite3.Error); ok {
			if driverErr.Code == sqlite3.ErrConstraint {
				return true
			}
		}
		return false
	} else if s.dbtype == "mysql" {
		// 1062 duplicate key
		//
		if driverErr, ok := err.(*mysql.MySQLError); ok {
			if driverErr.Number == 1062 {
				return true
			}
		}
		return false
	}
	return false
}

/* round 2
func (s *RouteStore) GetAllForUser(string) []routes.Route{ }
func (s *RouteStore) GetRecentlyAdded() []routes.Route{}
func (s *RouteStore) GetRecentlyModified() []routes.Route {}
*/
