gosqlcache
==========
 > A postgres caching driver on top of lib/pq that implements database/sql

How it works
------------
Gosqlcache registers itself as a driver and allows you to register query strings to be cached for a period of time. When you run a registered query, the driver will match the query and argument list against cached results. If a registered query is run and there are no cached results, it will cache them for you.

Exec statements and unregistered queries pass directly through to lib/pq.

Usage Example
-------------

```go
package main

import "log"
import "github.com/astaxie/beego/cache"
import "database/sql"
import "database/sql/driver"
import "os"
import "time"

import . ".."

// If your driver doesn't expose itself publicly,
// database.sql.DB offers a Driver() method to return the underlying driver
// but you have to Open() it first

import _ "github.com/lib/pq"

var pq driver.Driver

func init() {
	db, err := sql.Open("postgres", "")
	if err != nil {
		panic("can't get pq driver: " + err.Error())
	}
	pq = db.Driver()
}

func main() {
	memCache := SecondsTimeoutCacheWrapper{cache.NewMemoryCache()}
	logger := log.New(os.Stderr, "go-sql-cache: ", 0)
	sqlCacher := NewSqlCacher(&memCache, logger)
	sqlCacheDriver := NewSqlCacheDriver(pq, sqlCacher, logger)
	sqlCacheDriver.RegisterQuery("SELECT pg_sleep(1)", 1*time.Minute)
	sql.Register("go-sql-cache", sqlCacheDriver)
	db, err := sql.Open("go-sql-cache", "postgres://0.0.0.0:5432/postgres?sslmode=disable")
	if err != nil {
		panic("failed to connect to postgres: " + err.Error())
	}
	t1 := time.Now()
	_, err = db.Query("SELECT pg_sleep(1)")
	if err != nil {
		panic("initial query failed: " + err.Error())
	}
	logger.Println("First query took", time.Since(t1))
	t2 := time.Now()
	_, err = db.Query("SELECT pg_sleep(1)")
	if err != nil {
		panic("cached query failed: " + err.Error())
	}
	logger.Println("Second query took", time.Since(t2))
}
```

Warnings
--------
 - Might have unexpected results when using function calls or dynamic values in query strings. i.e. now()

TODOs
-----
 - add stats