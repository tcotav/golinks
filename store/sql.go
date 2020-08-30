package store

// we want a map with a map in it :D

var SQLDict map[string]map[string]string

func initSQLDict() {
	SQLDict = make(map[string]map[string]string)

	SQLDict["sqlite"] = map[string]string{
		"insertRoute":    "INSERT INTO routes(short_key, url, creatorid, teamid, created_at, modified_at, last_modified_by) VALUES (?,?,?,?,?,?,?)",
		"insertUser":     "INSERT INTO users(name, created_at, isadmin) VALUES(?,?,?)",
		"getUser":        "SELECT id, name, isadmin FROM users where name = ?",
		"getAllUsers":    "SELECT id, name, isadmin FROM users",
		"makeUserAdmin":  "UPDATE users SET isadmin = 1 where name = ?",
		"updateURLLock":  "UPDATE routes SET lock=1, last_modified_by=?, modified_at=? where short_key = ?",
		"getRouteSQL":    "SELECT short_key, url, creatorid, team, created_at, modified_at, last_modified_by FROM routes where short_key = ?",
		"getURLSQL":      "SELECT  url FROM routes where short_key = ?",
		"getURLIsLocked": "SELECT locked FROM routes where short_key = ?",
		"updateURLSQL":   "UPDATE routes SET url=?, last_modified_by=?, modified_at=? where short_key = ?",
		"deleteRouteSQL": "DELETE FROM routes where short_key = ?",
	}

	SQLDict["mysql"] = map[string]string{
		"insertRoute":    "INSERT INTO routes(short_key, url, creatorid, team, created_at, modified_at, last_modified_by) VALUES (?,?,?,?,?,?,?)",
		"insertUser":     "INSERT INTO users(name, created_at, isadmin) VALUES(?,?,?)",
		"getUser":        "SELECT id, name, isadmin FROM users where name = ?",
		"getAllUsers":    "SELECT id, name, isadmin FROM users",
		"makeUserAdmin":  "UPDATE users SET isadmin = 1, modified_at=?, last_modified_by=? where id = ?",
		"updateURLLock":  "UPDATE routes SET locked=1, last_modified_by=?, modified_at=? where short_key = ?",
		"getRouteSQL":    "SELECT short_key, url, created_at, creatorid, team, modified_at, last_modified_by FROM routes where short_key = ?",
		"getAllRoutes":   "SELECT short_key, url, creatorid, team, last_modified_by FROM routes",
		"getURLSQL":      "SELECT  url FROM routes where short_key = ?",
		"getURLIsLocked": "SELECT locked FROM routes where short_key = ?",
		"updateURLSQL":   `UPDATE routes SET url=?, last_modified_by=?, modified_at=DATE_FORMAT(?, "%Y-%m-%d %H:%i:%s") where short_key = ?`,
		"deleteRouteSQL": "DELETE FROM routes where short_key = ?",
	}
}

func GetSQL(dbType string, queryTag string) string {
	if SQLDict == nil {
		initSQLDict()
	}
	return SQLDict[dbType][queryTag]
}
