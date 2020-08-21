package store

import (
	"github.com/tcotav/golinks/routes"
)

type RouteStore interface {
	Add(routes.Route)
	Modify(routes.Route)
	Get(string) routes.Route
	Delete(string) bool
	//GetAllForUser(string) []routes.Route
	//GetRecentlyAdded() []routes.Route
	//GetRecentlyModified() []routes.Route
}
