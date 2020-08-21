package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	// database driver for sql package
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
const teamAuthHeader = "TeamNameAuth"

type MsgReturn struct {
	ReturnCode int
	Routes     []routes.Route
	Message    string
}

func add(w http.ResponseWriter, r *http.Request) {
	if authRequired {
		user := r.Header.Get(userAuthHeader)
		team := r.Header.Get(teamAuthHeader)
		if user == "" || team == "" {
			http.Error(w, "You must be authenticated", http.StatusInternalServerError)
			return
		}
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
	if s.IsSQLErrDuplicateContraint(err) {
		http.Error(w, "Key Exists", http.StatusBadRequest)
		return
	} else if err != nil {
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

func edit(w http.ResponseWriter, r *http.Request) {
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

func get(w http.ResponseWriter, r *http.Request) {
	// format /{secretname}
	vars := mux.Vars(r)
	shortKey, ok := vars["short_key"]
	if !ok {
		http.Error(w, "Invalid url format", http.StatusInternalServerError)
		return
	}

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

var s *store.SQLStore
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

	s, err = store.NewStore(useDB, database)
	if err != nil {
		// kill process because we won't have a DB anyway
		log.Fatal(err.Error())
	}

	r := mux.NewRouter()
	r.HandleFunc("/{short_key}", get)
	r.HandleFunc("/add/{secret}", add)
	//r.HandleFunc("//{secret}", edit)
	//r.HandleFunc("/s/{secret}", delete)
	srv := &http.Server{
		Handler:      r,
		Addr:         fmt.Sprintf("%s:%s", listenAddress, listenPort),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Print("Listening at:", listenAddress, ":", listenPort)
	log.Fatal(srv.ListenAndServe())
}
