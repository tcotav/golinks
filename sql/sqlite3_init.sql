CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, name TEXT, isadmin int, created_at datetime);
CREATE UNIQUE INDEX idx_users_name ON users(name);
CREATE TABLE IF NOT EXISTS routes (id INTEGER PRIMARY KEY, 
			short_key TEXT, 
			url TEXT,	
			creatorid int, 
			teamid int, 
			created_at datetime, 
			modified_at datetime, 
			last_modified_by int,
			locked int default 0, -- 0 means unlocked, 1 is locked
			FOREIGN KEY(creatorid) REFERENCES users(id),
			FOREIGN KEY(last_modified_by) REFERENCES users(id)
			);
CREATE UNIQUE INDEX idx_short_key ON routes(short_key);