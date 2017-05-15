package util

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"

	redis "gopkg.in/redis.v5"

	simplejson "github.com/bitly/go-simplejson"
)

func loadJSONConf(path string) (*simplejson.Json, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	js, err := simplejson.NewJson(buf)
	return js, err
}

//InitConf init redis and mysql connection with conf info
func InitConf(path string) (*redis.Client, *sql.DB) {
	js, err := loadJSONConf(path)
	if err != nil {
		log.Fatal(err)
	}
	redisHost, err := js.Get("redis").Get("host").String()
	if err != nil {
		log.Fatal(err)
	}
	kv := InitRedisHost(redisHost)
	dbAccess, err := js.Get("mysql").Get("access").String()
	if err != nil {
		log.Fatal(err)
	}
	dbHost, err := js.Get("mysql").Get("host").String()
	if err != nil {
		log.Fatal(err)
	}
	db, err := InitDBParam(dbAccess, dbHost)
	if err != nil {
		log.Fatal(err)
	}
	return kv, db
}
