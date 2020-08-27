package store

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
	"github.com/tcotav/golinks/routes"
)

const testDB string = "./testdb"

var (
	dbConn *sql.DB
)

func init() {
	// if db exists, remove it
	os.Remove(testDB)

	var err error
	// create a shared connection
	dbConn, err = sql.Open("sqlite3", testDB)
	if err != nil {
		log.Fatal(err.Error())
	}

	// really we could/should use a shared STORE for this whole thing, but...
	s, err := GetStore("sqlite", dbConn, nil, -1)
	if err != nil {
		// kill process because we won't have a DB anyway
		log.Fatal(err.Error())
	}
	// we do the setup of the database.  This probably wouldn't happen in RW code.
	initStatements := []string{
		"CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, created_at datetime)",
		"CREATE UNIQUE INDEX idx_users_name ON users(name)",
		`CREATE TABLE IF NOT EXISTS routes (id INTEGER PRIMARY KEY, 
			short_key TEXT, 
			url TEXT,	
			creatorid int, 
			teamid int, 
			created_at datetime, 
			modified_at datetime, 
			last_modified_by int,
			FOREIGN KEY(creatorid) REFERENCES users(id),
			FOREIGN KEY(last_modified_by) REFERENCES users(id)
			)`,
		"CREATE UNIQUE INDEX idx_short_key ON routes(short_key)",
	}
	for _, stmt := range initStatements {

		statement, err := dbConn.Prepare(stmt)
		if err != nil {
			// kill process because we won't have a DB anyway
			log.Fatal(err.Error())
		}
		_, err = statement.Exec()
		if err != nil {
			// kill process because we won't have a DB anyway
			log.Fatal(err.Error())
		}
	}
	r, err := routes.NewRoute("a", "http://www.google.com", "t@t.com", "team@t.com")
	if err != nil {
		log.Fatal(err)
	}

	i, err := s.Add(r)
	if err != nil && i != 1 {
		log.Fatal(err)
	}
}

func TestAddUser(t *testing.T) {
	s, err := GetStore("sqlite", dbConn, nil, -1)
	if err != nil {
		// kill process because we won't have a DB anyway
		t.Error(err.Error())
		return
	}
	id, err := s.GetUserID("t@t.com")
	if err != nil {
		t.Error(err.Error())
	}

	// same username again, should get same id
	id2, err := s.GetUserID("t@t.com")
	if err != nil {
		t.Error(err.Error())
	}

	if id != id2 {
		t.Error("User IDs don't match for GetUserID")
	}
}

func TestAdd(t *testing.T) {
	s, err := GetStore("sqlite", dbConn, nil, -1)
	if err != nil {
		// kill process because we won't have a DB anyway
		t.Error(err.Error())
		return
	}
	r, err := routes.NewRoute("d", "http://www.google.com", "t1@t.com", "te@t.com")
	if err != nil {
		t.Error(err.Error())
		return
	}

	i, err := s.Add(r)
	if err != nil && i != 1 {
		t.Error(err.Error())
		return
	}
	_, err = s.Add(r)
	if !s.IsSQLErrUniqueContraint(err) {
		t.Error(err.Error())
	}
}

func TestDelete(t *testing.T) {
	s, err := GetStore("sqlite", dbConn, nil, -1)
	if err != nil {
		// kill process because we won't have a DB anyway
		t.Error(err.Error())
		return
	}
	r, err := routes.NewRoute("q", "http://www.google.com", "t1@t.com", "te@t.com")
	if err != nil {
		t.Error(err.Error())
		return
	}

	i, err := s.Add(r)
	if err != nil && i != 1 {
		t.Error(err.Error())
		return
	}
	err = s.Delete("q")
	if err != nil {
		t.Error(err.Error())
	}
	r, err = s.Get("q")
	if err != nil {
		e := err.Error()
		if e != "No match found" {
			t.Error(err.Error())
		}
	}
}

func TestUpdate(t *testing.T) {
	s, err := GetStore("sqlite", dbConn, nil, -1)
	if err != nil {
		// kill process because we won't have a DB anyway
		t.Error(err.Error())
		return
	}
	r, err := s.Get("a")
	if err != nil {
		t.Error(err.Error())
		return
	}

	i, err := s.Modify(r)
	if err != nil {
		t.Error(err.Error())
		return
	}
	if i != 1 {
		t.Errorf("Unexpected rows modified -- expected 1 and got %d", i)
		return
	}
}

func TestGet(t *testing.T) {
	s, err := GetStore("sqlite", dbConn, nil, -1)
	if err != nil {
		// kill process because we won't have a DB anyway
		t.Error(err.Error())
		return
	}
	v, err := s.GetURL("a")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if v != "http://www.google.com" {
		t.Error("Expected: val1, got: ", v)
	}
}

func TestRedis(t *testing.T) {

	redisServer, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer redisServer.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: redisServer.Addr(),
	})

	// NewStore is too smart so we have to force it
	s, err := NewStore("sqlite", dbConn, redisClient, 300)
	if err != nil {
		// kill process because we won't have a DB anyway
		t.Error(err.Error())
		return
	}
	r, err := routes.NewRoute("x", "http://www.google.com", "t1@t.com", "te@t.com")
	if err != nil {
		t.Error(err.Error())
		return
	}
	i, err := s.Add(r)
	if err != nil && i != 1 {
		t.Error(err.Error())
		return
	}
	v, err := s.GetURL("x")
	if err != nil {
		t.Error(err.Error())
	}
	if v != "http://www.google.com" {
		t.Error("Expected: val1, got: ", v)
	}

	v2 := "http://www.t.com"
	r.URL = v2
	i, err = s.Modify(r)
	if err != nil {
		t.Error(err.Error())
	}
	if i != 1 {
		t.Errorf("Unexpected rows modified -- expected 1 and got %d", i)
	}
	v, err = s.GetURL("x")
	if err != nil {
		t.Error(err.Error())
	}
	if v != v2 {
		t.Error("Expected does not match found : ", v2, v)
	}
	r, err = s.Get("x")
	if err != nil {
		t.Error(err.Error())
	}
	if r.URL != v2 {
		t.Error("Expected does not match found : ", v2, r.URL)
	}

	err = s.Delete("x")
	if err != nil {
		t.Error(err.Error())
	}
	r, err = s.Get("x")
	if err != nil {
		e := err.Error()
		if e != "No match found" {
			t.Error(err.Error())
		}
	}
}

/*
func TestUpdate(t *testing.T) {
	s, err := NewStore(dbConn)
	if err != nil {
		// kill process because we won't have a DB anyway
		t.Error(err.Error())
		return
	}
	i, err := s.Update("test1", "val999")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if i != 1 {
		t.Error("expected: change count 1, got: ", i)
	}
}

func TestCreate(t *testing.T) {
	s, err := NewStore(dbConn)
	if err != nil {
		// kill process because we won't have a DB anyway
		t.Error(err.Error())
		return
	}
	i, err := s.Create("newkey", "newval")
	if err != nil {
		t.Error(err.Error())
		return
	}

	if i != 1 {
		t.Error("expected: change count 1, got: ", i)
	}
}

func TestDelete(t *testing.T) {
	s, err := NewStore(dbConn)
	if err != nil {
		// kill process because we won't have a DB anyway
		t.Error(err.Error())
		return
	}
	i, err := s.Delete("test1")
	if err != nil {
		t.Error(err.Error())
		return
	}

	if i != 1 {
		t.Error("expected: change count 1, got: ", i)
	}
}
*/
