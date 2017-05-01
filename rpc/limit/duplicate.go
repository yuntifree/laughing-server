package main

import (
	"laughing-server/proto/limit"
	"log"
	"strconv"
	"time"

	redis "gopkg.in/redis.v5"
)

const (
	nonceSet = "limit:nonce:set"
	imeiSet  = "limit:imei:set"
	interval = 60 * 10
)

func addKeyId(kv *redis.Client, key, id string, ts int64) {
	_, err := kv.ZAdd(key, redis.Z{Member: id, Score: float64(ts)}).Result()
	if err != nil {
		log.Printf("addKeyId ZAdd failed:%s %s %v", key, id, err)
	}
	_, err = kv.ZRemRangeByScore(key, "0", strconv.Itoa(int(ts-interval))).Result()
	if err != nil {
		log.Printf("checkDuplicate ZRemRangeByScore failed:%s %s %v", key, id, err)
	}
}

func checkDuplicate(kv *redis.Client, ctype limit.CheckType, id string) bool {
	var key string
	switch ctype {
	case limit.CheckType_NONCE:
		key = nonceSet
	case limit.CheckType_IMEI:
		key = imeiSet
	default:
		return false
	}
	ts := time.Now().Unix()
	s, err := kv.ZScore(key, id).Result()
	if err != nil {
		if err == redis.Nil {
			log.Printf("key:%s id:%s not found", key, id)
			addKeyId(kv, key, id, ts)
			return false
		}
		log.Printf("checkDuplicate ZScore failed:%s %s %v", key, id, err)
		return false
	}
	if s > 0 {
		return true
	}
	addKeyId(kv, key, id, ts)
	return false
}
