package util

import (
	"strconv"
	"time"

	redis "gopkg.in/redis.v5"
)

const (
	redisHost   = "r-wz9191666aa18664.redis.rds.aliyuncs.com:6379"
	redisPasswd = "YXZHwifiredis01server"
)

//InitRedis return initialed redis client
func InitRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: redisPasswd,
		DB:       0,
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
