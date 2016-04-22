package gosqlcache

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
	"os"
	"testing"
	"time"
)

// docker run -p 6379 -d redis
// export REDIS_HOST=192.168.99.100
// export REDIS_PORT=32768
// export REDIS_USER=
// export REDIS_PASS=

// docker run -p 5432 -d postgres
// export PG_HOST=192.168.99.100
// export PG_PORT=32771
// export PG_DB=postgres
// export PG_USER=postgres
// export PG_PASS=

func CreateTestTable(db *sql.DB) (sql.Result, error) {
	return db.Exec("create table if not exists test (id serial, name varchar(63), age int, value float, is_awesome bool, created timestamp)")
}

func TestRedisNoCache(t *testing.T) {
	bm, err := cache.NewCache("redis", fmt.Sprintf(`{"conn":"%s:%s"}`, os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")))
	if err != nil {
		t.Fatal("Could not connect to redis: ", err)
	}
	t.Log("Established redis connection")
	redisCache := SecondsTimeoutCacheWrapper{bm}
	db, err := sql.Open("postgres-cached", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("PG_USER"), os.Getenv("PG_PASS"), os.Getenv("PG_HOST"), os.Getenv("PG_PORT"), os.Getenv("PG_DB")))
	if err != nil {
		t.Fatal("Could not connect to database", err)
	}
	t.Log("Established database connection")
	defer db.Close()
	SetCacher(&redisCache)
	defer _cache.ClearAll()
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
}

func TestRedisCacheRegister(t *testing.T) {
	bm, err := cache.NewCache("redis", fmt.Sprintf(`{"conn":"%s:%s"}`, os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")))
	if err != nil {
		t.Fatal("Could not connect to redis: ", err)
	}
	redisCache := SecondsTimeoutCacheWrapper{bm}
	db, err := sql.Open("postgres-cached", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("PG_USER"), os.Getenv("PG_PASS"), os.Getenv("PG_HOST"), os.Getenv("PG_PORT"), os.Getenv("PG_DB")))
	if err != nil {
		t.Fatal("Could not connect to database")
	}
	defer db.Close()
	SetCacher(&redisCache)
	defer _cache.ClearAll()
	RegisterQuery("SELECT 5", 1*time.Minute)
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
	err = rows.Scan(&n)
	if err != nil {
		t.Fatal("Could not read result: ", err)
	}
	if n != 5 {
		t.Fatal("Bad result: ", n)
	}
}

func TestRedisCacheRead(t *testing.T) {
	bm, err := cache.NewCache("redis", fmt.Sprintf(`{"conn":"%s:%s"}`, os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")))
	if err != nil {
		t.Fatal("Could not connect to redis: ", err)
	}
	redisCache := SecondsTimeoutCacheWrapper{bm}
	db, err := sql.Open("postgres-cached", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("PG_USER"), os.Getenv("PG_PASS"), os.Getenv("PG_HOST"), os.Getenv("PG_PORT"), os.Getenv("PG_DB")))
	if err != nil {
		t.Fatal("Could not connect to database")
	}
	defer db.Close()
	SetCacher(&redisCache)
	defer _cache.ClearAll()
	_cache.PutQueryRows("SELECT 5", []driver.Value{}, &cachedRows{
		cols: []string{"five"},
		data: [][]driver.Value{[]driver.Value{3}},
	}, 1*time.Minute)
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
	err = rows.Scan(&n)
	if err != nil {
		t.Fatal("Could not read result: ", err)
	}
	if n != 3 {
		t.Fatal("Bad result: ", n)
	}
}

func TestRedisCacheWrite(t *testing.T) {
	bm, err := cache.NewCache("redis", fmt.Sprintf(`{"conn":"%s:%s"}`, os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")))
	if err != nil {
		t.Fatal("Could not connect to redis: ", err)
	}
	redisCache := SecondsTimeoutCacheWrapper{bm}
	db, err := sql.Open("postgres-cached", fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", os.Getenv("PG_USER"), os.Getenv("PG_PASS"), os.Getenv("PG_HOST"), os.Getenv("PG_PORT"), os.Getenv("PG_DB")))
	if err != nil {
		t.Fatal("Could not connect to database")
	}
	defer db.Close()
	SetCacher(&redisCache)
	defer _cache.ClearAll()
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
	err = rows.Scan(&n)
	if err != nil {
		t.Fatal("Could not read result: ", err)
	}
	if n != 5 {
		t.Fatal("Bad result: ", n)
	}
	cachedval := _cache.GetQueryRows("SELECT 5", []driver.Value{})
	// t.Logf("Read from cache directly: %#v", cachedval) //[]byte{0xa, 0xff, 0x81, 0x5, 0x1, 0x2, 0xff, 0x84, 0x0, 0x0, 0x0, 0xff, 0x86, 0xff, 0x82, 0x0, 0xff, 0x81, 0xe, 0xff, 0x85, 0x4, 0x1, 0x2, 0xff, 0x86, 0x0, 0x1, 0xc, 0x1, 0x10, 0x0, 0x0, 0x21, 0xff, 0x86, 0x0, 0x2, 0x7, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x73, 0x8, 0x5b, 0x5d, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xff, 0x87, 0x2, 0x1, 0x2, 0xff, 0x88, 0x0, 0x1, 0xc, 0x0, 0x0, 0x31, 0xff, 0x88, 0xb, 0x0, 0x1, 0x8, 0x3f, 0x63, 0x6f, 0x6c, 0x75, 0x6d, 0x6e, 0x3f, 0x4, 0x64, 0x61, 0x74, 0x61, 0x10, 0x5b, 0x5d, 0x5b, 0x5d, 0x64, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e, 0x56, 0x61, 0x6c, 0x75, 0x65, 0xff, 0x8b, 0x2, 0x1, 0x2, 0xff, 0x8c, 0x0, 0x1, 0xff, 0x8a, 0x0, 0x0, 0xc, 0xff, 0x89, 0x2, 0x1, 0x2, 0xff, 0x8a, 0x0, 0x1, 0x10, 0x0, 0x0, 0x10, 0xff, 0x8c, 0xd, 0x0, 0x1, 0x1, 0x5, 0x69, 0x6e, 0x74, 0x36, 0x34, 0x4, 0x2, 0x0, 0xa}
	// t.Logf("%s", string(append([]byte("\n"), cachedval.([]byte)[16:]...)))
	// var cr = cachedRows{}
	// err = cr.GobDecode(cachedval.([]byte))
	// err = cr.GobDecode([]byte("????\n!??column[]string????\n-??fivedata[][]driver.Value??????\n??????\nint\n"))
	if err != nil {
		t.Fatal("Couldn't decode cached result", err)
	}
	if cachedval.(*cachedRows).data[0][0] /*.(int64)*/ != int64(5) {
		t.Fatalf("Cache doesn't contain 5, got %[1]T : %#[1]v instead", cachedval.(*cachedRows).data[0][0])
	}
	if int64(5) != 5 {
		t.Fatalf("wtf")
	}
}
