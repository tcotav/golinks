package store

import (
	"github.com/tcotav/golinks/routes"
)

type RouteStore interface {
	Add(routes.Route) (int, error)
	Modify(routes.Route) (int, error)
	Get(string) routes.Route
	Delete(string) bool
	GetUserID(username string) (int, error)
	GetURL(string) (string, error)
	Lock(routes.Route) (bool, error)
	//GetAllForUser(string) []routes.Route
	//GetRecentlyAdded() []routes.Route
	//GetRecentlyModified() []routes.Route
}
