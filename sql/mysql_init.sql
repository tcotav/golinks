CREATE DATABASE routes;
USE routes;

CREATE TABLE IF NOT EXISTS users (id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, name VARCHAR(50), created_at datetime, modified_at datetime, last_modified_by int);
CREATE UNIQUE INDEX idx_users_name ON users(name);
CREATE TABLE IF NOT EXISTS routes (id INTEGER NOT NULL AUTO_INCREMENT PRIMARY KEY, 
			short_key VARCHAR(20), 
			url VARCHAR(4000),	 -- this seems ridiculous but you never know
			creatorid int, 
			team VARCHAR(40), 
			created_at datetime, 
			modified_at datetime, 
			last_modified_by int,
			locked int default 0, -- 0 means unlocked, 1 is locked
			FOREIGN KEY(creatorid) REFERENCES users(id),
			FOREIGN KEY(last_modified_by) REFERENCES users(id)
			);
CREATE UNIQUE INDEX idx_short_key ON routes(short_key);