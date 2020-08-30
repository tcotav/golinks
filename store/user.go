package store

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	IsAdmin int    `json:"isadmin,omitempty"`
}
