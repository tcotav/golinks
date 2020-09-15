package store

import (
	"database/sql"
	"errors"
	"fmt"
	"sync"
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
	init     sync.Once
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

func (s *DataStore) GetUser(username string) (*User, error) {
	// create or get
	rows, err := s.db.Query(GetSQL(s.dbtype, "getUser"), username)
	if err != nil {
		return &User{}, err
	}
	defer rows.Close()

	var user User
	for rows.Next() {
		err := rows.Scan(&user.ID, &user.Name, &user.IsAdmin)
		if err != nil {
			return &User{}, err
		}
		return &user, nil
	}

	now := time.Now()
	stmt, err := s.db.Prepare(GetSQL(s.dbtype, "insertUser"))
	if err != nil {
		return &User{}, err
	}
	res, err := stmt.Exec(username, now, 0)
	if err != nil {
		return &User{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return &User{}, err
	}
	return &User{ID: int(id), Name: username}, nil

}

func (s *DataStore) DumpAllRoutes() ([]routes.Route, error) {
	rows, err := s.db.Query(GetSQL(s.dbtype, "getAllRoutes"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var r routes.Route
	routeList := make([]routes.Route, 0)
	for rows.Next() {
		err := rows.Scan(&r.ShortKey, &r.URL, &r.Creator, &r.Team, &r.LastModifiedBy)
		if err != nil {
			return nil, err
		}
		routeList = append(routeList, r)
	}
	return routeList, nil
}

func (s *DataStore) GetAllUsers() error {
	rows, err := s.db.Query(GetSQL(s.dbtype, "getAllUsers"))
	if err != nil {
		return err
	}
	defer rows.Close()

	var user User
	for rows.Next() {
		err := rows.Scan(&user.ID, &user.Name, &user.IsAdmin)
		if err != nil {
			return err
		}
		fmt.Println(user)
	}
	return nil
}

func (s *DataStore) MakeAdmin(username string, admin string) (int, error) {
	// is admin authorized to do this?
	adminUser, err := s.GetUser(admin)
	if err != nil {
		return -1, err
	}
	if adminUser.IsAdmin != 1 {
		return -1, fmt.Errorf("%s is not an admin, makeadmin called failed.", admin)
	}

	u, err := s.GetUser(username)
	if err != nil {
		return -1, err
	}
	now := time.Now().Format(routes.TimeFormat)
	res, err := s.db.Exec(GetSQL(s.dbtype, "makeUserAdmin"), now, adminUser.ID, u.ID)
	if err != nil {
		return -1, err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return int(affect), nil
}

// Lock locks the entry so that it requires admin to unlock and change
func (s *DataStore) Lock(r routes.Route) (int, error) {
	now := time.Now()
	user, err := s.GetUser(r.LastModifiedBy)
	if err != nil {
		return -1, err
	}
	if user.IsAdmin != 1 {
		return -1, fmt.Errorf("User %s is not admin", user.Name)
	}
	res, err := s.db.Exec(GetSQL(s.dbtype, "updateURLLock"), user.ID, now, r.ShortKey)
	if err != nil {
		return -1, err
	}
	affect, err := res.RowsAffected()
	if err != nil {
		return -1, err
	}
	return int(affect), nil
}

func (s *DataStore) Add(r routes.Route) (int, error) {
	stmt, err := s.db.Prepare(GetSQL(s.dbtype, "insertRoute"))
	if err != nil {
		return -1, err
	}
	u, err := s.GetUser(r.Creator)
	if err != nil {
		return -1, err
	}
	res, err := stmt.Exec(r.ShortKey, r.URL, u.ID, r.Team, r.CreatedAt, r.ModifiedAt, u.ID)
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

// Get is
func (s *DataStore) Get(k string) (routes.Route, error) {
	rows, err := s.db.Query(GetSQL(s.dbtype, "getRouteSQL"), k)
	if err != nil {
		return routes.Route{}, err
	}
	defer rows.Close()

	var r routes.Route
	for rows.Next() {
		err := rows.Scan(&r.ShortKey, &r.URL, &r.CreatedAt, &r.Creator, &r.Team, &r.ModifiedAt, &r.LastModifiedBy)
		if err != nil {
			return routes.Route{}, err
		}
		return r, nil
	}
	return routes.Route{}, errors.New("No match found")
}

// GetURL is the main entry point for the app.  It is the shortlink.  The other getters are
// designed for potentially slower access via an editor tool.
func (s *DataStore) GetURL(k string) (string, error) {
	// we should not use the local cache if we have redis as that assumes we have multiple nodes
	// and could have a local cache go stale
	if s.redis != nil {
		sRet, _ := s.redis.Get(k).Result()
		if sRet != "" {
			// update lcl cache and return
			return sRet, nil
		}
	} else if s.cache != nil { // check local cache for single node runs
		v, ok := s.cache.Get(k)
		if ok {
			return v.(string), nil
		}
	}

	// then database
	rows, err := s.db.Query(GetSQL(s.dbtype, "getURLSQL"), k)
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

		if s.redis != nil {
			// update redis cache
			s.redis.Set(k, url, s.redisTTL).Err()
		} else if s.cache != nil {
			// update local cache
			s.cache.Add(k, url)
		}
		return url, nil
	}
	return "", errors.New("No match found")
}

func (s *DataStore) Modify(r routes.Route) (int, error) {
	// what about case where we are changing the shortkey -- how to invalidate caches?
	now := time.Now().Format(routes.TimeFormat)
	user, err := s.GetUser(r.LastModifiedBy)
	if err != nil {
		return -1, err
	}

	// double check the lock status of the key in question before we move on
	// then database
	rows, err := s.db.Query(GetSQL(s.dbtype, "getURLIsLocked"), r.ShortKey)
	if err != nil {
		return -1, err
	}
	defer rows.Close()
	var isLocked int
	for rows.Next() {
		err := rows.Scan(&isLocked)
		if err != nil {
			return -1, err
		}
	}

	// exit if not allowed in
	if isLocked == 1 && user.IsAdmin != 1 {
		return -1, fmt.Errorf("User %s is not admin", user.Name)
	}

	// then move on
	res, err := s.db.Exec(GetSQL(s.dbtype, "updateURLSQL"), r.URL, user.ID, now, r.ShortKey)
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
	if s.redis != nil {
		s.redis.Set(r.ShortKey, r.URL, s.redisTTL).Err()
	} else if s.cache != nil {
		if _, ok := s.cache.Get(r.ShortKey); !ok {
			s.cache.Remove(r.ShortKey)
			s.cache.Add(r.ShortKey, r.URL)
		}
	}
	return int(affect), nil
}

func (s *DataStore) Delete(k string) error {
	// what about case where we are changing the shortkey -- how to invalidate caches?
	res, err := s.db.Exec(GetSQL(s.dbtype, "deleteRouteSQL"), k)
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

	if s.redis != nil {
		s.redis.Del(k).Err()
	} else if s.cache != nil {
		if _, ok := s.cache.Get(k); !ok {
			s.cache.Remove(k)
		}
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
			if driverErr.Number == 1169 || driverErr.Number == 1062 {
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
