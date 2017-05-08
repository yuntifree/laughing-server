package main

import (
	"database/sql"
	"log"
	"strconv"

	redis "gopkg.in/redis.v5"
)

const (
	userTokenSet = "user:token"
)

func setCachedToken(kv *redis.Client, uid int64, token string) {
	_, err := kv.HSet(userTokenSet, strconv.Itoa(int(uid)), token).Result()
	if err != nil {
		log.Printf("setCacheToken failed:%v", err)
	}
}

func getCachedToken(kv *redis.Client, uid int64) (string, error) {
	res, err := kv.HGet(userTokenSet, strconv.Itoa(int(uid))).Result()
	if err != nil {
		log.Printf("getCachedToken failed:%d %v", uid, err)
		return "", err
	}
	return res, nil
}

func checkToken(db *sql.DB, kv *redis.Client, uid int64, token string) bool {
	etoken, err := getCachedToken(kv, uid)
	if err == nil && etoken == token {
		return true
	}
	err = db.QueryRow("SELECT token FROM users WHERE uid = ?", uid).Scan(&etoken)
	if err != nil {
		return false
	}
	setCachedToken(kv, uid, etoken)
	if token == etoken {
		return true
	}
	return false
}

func checkBackToken(db *sql.DB, uid int64, token string) bool {
	var etoken string
	err := db.QueryRow("SELECT token FROM back_login WHERE uid = ?", uid).
		Scan(&etoken)
	if err != nil {
		return false
	}
	if token == etoken {
		return true
	}
	return false
}
