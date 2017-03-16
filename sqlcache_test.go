package gosqlcache

import (
	"database/sql"
	"database/sql/driver"
	"github.com/astaxie/beego/cache"
	"testing"
	"time"
)

import . "./test"

func TestNoCache(t *testing.T) {

	// Start Setup
	memCache := SecondsTimeoutCacheWrapper{cache.NewMemoryCache()}
	log := NewTestLogger(t)
	sqlCacher := NewSqlCacher(&memCache, log)
	sqlCacheSpy := SpyOnSqlCacher(sqlCacher)
	mockConn := MockConn{}
	mockDriver := MockDriver{
		Conn: &mockConn,
	}
	sqlCacheDriver := NewSqlCacheDriver(&mockDriver, sqlCacheSpy, log)
	sql.Register(t.Name(), sqlCacheDriver)
	db, err := sql.Open(t.Name(), "")
	if err != nil {
		t.Fatal("Could not connect to database", err)
	}
	// End Setup

	defer db.Close()
	defer sqlCacher.ClearAll()

	mockConn.Rows = SpyOnRows(&cachedRows{
		cols: []string{"number"},
		data: [][]driver.Value{[]driver.Value{5}},
	})

	rows, err := db.Query("SELECT 5")
	if err != nil {
		t.Fatal("Could not select 5", err)
	}
	defer rows.Close()

	var n int
	ok := rows.Next()
	if !ok {
		t.Fatal("Could not get next")
	}
	rows.Scan(&n)
	if n != 5 {
		t.Fatal("Could not read result: ", n)
	}

	if sqlCacheSpy.WasEverCalled() {
		t.Fatal("Cache should not have been called without registering query")
	}

	if !mockConn.WasCalledWith("Query", []interface{}{"SELECT 5", []driver.Value{}}) {
		t.Logf("%#v", mockConn.Spy.Calls)
		t.Fatal("Query should have been called")
	}
}

func TestCacheCheckMiss(t *testing.T) {

	// Start Setup
	memCache := SecondsTimeoutCacheWrapper{cache.NewMemoryCache()}
	log := NewTestLogger(t)
	sqlCacher := NewSqlCacher(&memCache, log)
	sqlCacheSpy := SpyOnSqlCacher(sqlCacher)
	mockConn := MockConn{}
	mockDriver := MockDriver{
		Conn: &mockConn,
	}
	sqlCacheDriver := NewSqlCacheDriver(&mockDriver, sqlCacheSpy, log)
	sql.Register(t.Name(), sqlCacheDriver)
	db, err := sql.Open(t.Name(), "")
	if err != nil {
		t.Fatal("Could not connect to database", err)
	}
	// End Setup

	defer db.Close()
	defer sqlCacher.ClearAll()

	sqlCacheDriver.RegisterQuery("SELECT 5", 1*time.Minute)

	mockConn.Rows = SpyOnRows(&cachedRows{
		cols: []string{"number"},
		data: [][]driver.Value{[]driver.Value{5}},
	})

	rows, err := db.Query("SELECT 5")
	if err != nil {
		t.Fatal("Could not select 5", err)
	}
	defer rows.Close()

	var n int
	ok := rows.Next()
	if !ok {
		t.Fatal("Could not get next")
	}

	rows.Scan(&n)
	if n != 5 {
		t.Fatal("Could not read result: ", n)
	}

	if !sqlCacheSpy.WasCalled("GetQueryRows") {
		t.Fatal("SqlCache GetQueryRows should have been checked")
	}

	if !mockConn.WasCalled("Query") {
		t.Logf("%#v", mockConn.Spy.Calls)
		t.Fatal("Query should have been called since cache was empty")
	}

	if !sqlCacheSpy.WasCalled("PutQueryRows") {
		t.Logf("%#v", mockConn.Spy.Calls)
		t.Fatal("Query results should have been cached")
	}
}

func TestCacheCheckHit(t *testing.T) {

	// Start Setup
	memCache := SecondsTimeoutCacheWrapper{cache.NewMemoryCache()}
	log := NewTestLogger(t)
	sqlCacher := NewSqlCacher(&memCache, log)
	sqlCacheSpy := SpyOnSqlCacher(sqlCacher)
	mockConn := MockConn{}
	mockDriver := MockDriver{
		Conn: &mockConn,
	}
	sqlCacheDriver := NewSqlCacheDriver(&mockDriver, sqlCacheSpy, log)
	sql.Register(t.Name(), sqlCacheDriver)
	db, err := sql.Open(t.Name(), "")
	if err != nil {
		t.Fatal("Could not connect to database", err)
	}
	// End Setup

	defer db.Close()
	defer sqlCacher.ClearAll()

	sqlCacheDriver.RegisterQuery("SELECT 5", 1*time.Minute)

	cacheRows := SpyOnRows(&cachedRows{
		cols: []string{"number"},
		data: [][]driver.Value{[]driver.Value{5}},
	})
	err = sqlCacher.PutQueryRows("SELECT 5", []driver.Value{}, cacheRows, 1*time.Minute)
	if err != nil {
		t.Fatal("Unable to precache rows for test", err)
	}

	rows, err := db.Query("SELECT 5")
	if err != nil {
		t.Fatal("Could not select 5", err)
	}
	defer rows.Close()

	var n int
	ok := rows.Next()
	if !ok {
		t.Fatal("Could not get next")
	}

	rows.Scan(&n)
	if n != 5 {
		t.Fatal("Could not read result: ", n)
	}

	if !sqlCacheSpy.WasCalled("GetQueryRows") {
		t.Fatal("SqlCache GetQueryRows should have been checked")
	}

	if mockConn.WasCalled("Query") {
		t.Logf("%#v", mockConn.Spy.Calls)
		t.Fatal("Query should not have been called since cache was not empty")
	}

	if sqlCacheSpy.WasCalled("PutQueryRows") {
		t.Logf("%#v", mockConn.Spy.Calls)
		t.Fatal("Query results should not have been re-cached")
	}
}
