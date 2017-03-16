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
