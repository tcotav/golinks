## Go Short Url Service (GSUS)

### Overview

AuthN is 

#### Data structure:
- short_key - key used in url for lookup 
- url - returned url correlated to the key
- creator - creator of the k+v
- team - team of the creator
- created_at - date of creation
- modified_at - date last modified
- last_modified_by - user last modified key


#### Functionality

Requires access to DNS with default 
Primary function -  heavy read/lookup 

Editorial functions 
- add - who, what, when
- modify - who, what, when + (optional copy of previous or an audit trail)
- delete - who, what, when

Audit log - optional - logs of editorial activity
Audit tables - optional - data store (would require additional storage OR could be stored in flatfile/sqlite?) 
Traffic log - optional - clickstream? so we can generate a most popular lookups and show traffic info

Authentication and Authorization
- abstracted authentication
- allow everyone write access?  or set that at the start as everyone but have the placeholder in place to toggle?
- default to a file or sqlite

#### Scale Considerations

- initial version 
    - redis for multi nodes
    - audit log output
    - metrics output to log/prometheus

- small version/self-contained/personal
    - hashmap + localfile for storage.

- large version 
    - assume multiple front end nodes for read API
    - redis for common cache
    - mysql for audit tables and users

### Version 1 - standalone

- SQLite for data store
    - sudo apt install sqlite3

- Stubbed out hook + config for auth

