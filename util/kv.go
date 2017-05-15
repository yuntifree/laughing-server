package util

import (
	"strconv"
	"time"

	redis "gopkg.in/redis.v5"
)

const (
	redisHost = "10.11.121.205:6379"
)

//InitRedis return initialed redis client
func InitRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisHost,
		DB:   0,
	})
}

//InitRedisHost return initialed redis client with host
func InitRedisHost(host string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: host,
		DB:   0,
	})
}

//Report add address to server list
func Report(client *redis.Client, name, port string) {
	ip := GetInnerIP()
	addr := ip + port
	if ip == DebugHost {
		name += ":debug"
	}
	ts := time.Now().Unix()
	client.ZAdd(name, redis.Z{Member: addr, Score: float64(ts)})
	client.ZRemRangeByScore(name, "0", strconv.Itoa(int(ts-20)))
}

//ReportHandler handle report address
func ReportHandler(kv *redis.Client, name, port string) {
	for {
		time.Sleep(time.Second * 2)
		Report(kv, name, port)
	}
}
