package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	// database driver for sql package
	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/tcotav/golinks/routes"
	"github.com/tcotav/golinks/store"
)

func logLine(rPointer []byte, remoteAddr string, requestURI string,
	response string, respCode int) {

	// add metrics in here too?
	log.Printf("[%p] %d %s %s %s", rPointer, respCode, remoteAddr, requestURI, response)
}

const userAuthHeader = "UserNameAuth"

type MsgReturn struct {
	ReturnCode int
	Routes     []routes.Route
	Message    string
}

func doAuthCheck(r *http.Request) bool {
	if authRequired {
		user := r.Header.Get(userAuthHeader)
		if user == "" {
			return false
		}
	}
	return true
}

func edit(w http.ResponseWriter, r *http.Request) {
	user := r.Header.Get(userAuthHeader)

	if !doAuthCheck(r) {
		http.Error(w, "You must be authenticated", http.StatusInternalServerError)
		return
	}
	var route routes.Route
	// what to do in the case of conflict?
	err := json.NewDecoder(r.Body).Decode(&route)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	route.Creator = user
	route.LastModifiedBy = user

	// process and handle
	_, err = s.Add(route)
	// check if err is a duplicate constraint
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")
	// return both lists
	resp, _ := json.Marshal(MsgReturn{ReturnCode: http.StatusOK})
	w.Write(resp)

}

func add(w http.ResponseWriter, r *http.Request) {
	if !doAuthCheck(r) {
		http.Error(w, "You must be authenticated", http.StatusInternalServerError)
		return
	}
	var route routes.Route
	// what to do in the case of conflict?
	err := json.NewDecoder(r.Body).Decode(&route)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// process and handle
	_, err = s.Add(route)
	// check if err is a duplicate constraint
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "application/json")
	// return both lists
	resp, _ := json.Marshal(MsgReturn{ReturnCode: http.StatusOK})
	w.Write(resp)
}

/*
func getAllForUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortKey, ok := vars["short_key"]
	if !ok {
		http.Error(w, "Invalid url format", http.StatusInternalServerError)
		return
	}

}

func getAllForTeam(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortKey, ok := vars["short_key"]
	if !ok {
		http.Error(w, "Invalid url format", http.StatusInternalServerError)
		return
	}

}



func delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortKey, ok := vars["short_key"]
	if !ok {
		http.Error(w, "Invalid url format", http.StatusInternalServerError)
		return
	}

}
*/

const randomStr = "random"

//
// get is the main function -- responding to http://go/<key> with a redirect to the desired page
func get(w http.ResponseWriter, r *http.Request) {
	// format /{secretname}
	vars := mux.Vars(r)
	shortKey, ok := vars["short_key"]
	if !ok {
		http.Error(w, "Invalid url format", http.StatusInternalServerError)
		return
	}

	// easter egg -- shortKey == random
	URL, err := s.GetURL(shortKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		logLine([]byte(fmt.Sprintf("%p", r)), r.RemoteAddr, r.RequestURI, URL, http.StatusNotFound)
		return
	}

	// TODO need to send a redirect here
	logLine([]byte(fmt.Sprintf("%p", r)), r.RemoteAddr, r.RequestURI, URL, http.StatusOK)
	http.Redirect(w, r, URL, http.StatusFound)
}

func delete(w http.ResponseWriter, r *http.Request) {
	// format /{secretname}
	vars := mux.Vars(r)
	shortKey, ok := vars["short_key"]
	if !ok {
		http.Error(w, "Invalid url format", http.StatusInternalServerError)
		return
	}

	// easter egg -- shortKey == random
	err := s.Delete(shortKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		logLine([]byte(fmt.Sprintf("%p", r)), r.RemoteAddr, r.RequestURI, shortKey, http.StatusNotFound)
		return
	}

	// TODO need to send a redirect here
	logLine([]byte(fmt.Sprintf("%p", r)), r.RemoteAddr, r.RequestURI, shortKey, http.StatusOK)

	w.WriteHeader(http.StatusOK)
	// return both lists
	w.Write([]byte("OK"))
}

var s *store.DataStore
var authRequired bool

func main() {
	var err error
	viper.SetConfigName("config")         // name of config file (without extension)
	viper.AddConfigPath("/etc/golinks/")  // path to look for the config file in
	viper.AddConfigPath("$HOME/.golinks") // call multiple times to add many search paths
	viper.AddConfigPath(".")              // optionally look for config in the working directory
	viper.SetDefault("listenaddress", "127.0.0.1")
	viper.SetDefault("listenport", "8991")
	viper.SetDefault("authrequired", true)
	viper.SetDefault("datastore.use", "sqlite")
	viper.SetDefault("datastore.sqlite.drivername", "sqlite3")
	viper.SetDefault("datastore.sqlite.path", "./testdb")
	viper.SetDefault("cache.redis.ttl", 21600)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Fatal("Could not find config file")
		} else {
			// Config file was found but another error was produced
		}
	}

	listenAddress := viper.GetString("listenaddress")
	listenPort := viper.GetString("listenport")
	authRequired = viper.GetBool("authrequired")
	useDB := viper.GetString("datastore.use")

	var database *sql.DB

	if useDB == "sqlite" {
		driver := viper.GetString("datastore.sqlite.drivername")
		path := viper.GetString("datastore.sqlite.path")
		// initialize the store
		database, err = sql.Open(driver, path)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else if useDB == "mysql" {
		driver := viper.GetString("datastore.mysql.drivername")
		url := viper.GetString("datastore.mysql.url")
		database, err = sql.Open(driver, url)
		if err != nil {
			log.Fatal(err.Error())
		}
	} else {
		log.Fatal("Check config -- unknown db type set")
	}

	cacheType := viper.GetString("cache.use")
	ttl := -1

	var redisClient *redis.Client
	if cacheType == "remote" {
		ttl = viper.GetInt("cache.redis.ttl")
		redisClient = redis.NewClient(&redis.Options{
			Addr:     viper.GetString("cache.redis.host"),
			Password: viper.GetString("cache.redis.password"),
		})
	}

	s, err = store.GetStore(useDB, database, redisClient, ttl)
	if err != nil {
		// kill process because we won't have a DB anyway
		log.Fatal(err.Error())
	}

	r := mux.NewRouter()
	r.HandleFunc("/{short_key}", get)
	r.HandleFunc("/add/{secret}", add)
	r.HandleFunc("/edit/{secret}", edit)
	r.HandleFunc("/delete/{secret}", delete)
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", listenAddress, listenPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Print("Listening at:", listenAddress, ":", listenPort)
	log.Fatal(srv.ListenAndServe())
}
