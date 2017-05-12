package main

import (
	"database/sql"
	"laughing-server/ucloud"
	"laughing-server/util"
	"log"
	"net"
	"strings"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	nsq "github.com/nsqio/go-nsq"
)

func extractUid(msg *nsq.Message) int64 {
	js, err := simplejson.NewJson(msg.Body)
	if err != nil {
		log.Printf("HandlerMessage parse body failed:%s %v",
			string(msg.Body), err)
		return 0
	}
	uid, err := js.Get("uid").Int64()
	if err != nil {
		log.Printf("get uid failed:%s %v", string(msg.Body), err)
		return 0
	}
	return uid
}

func getUserHead(db *sql.DB, uid int64) string {
	var head string
	err := db.QueryRow("SELECT headurl FROM users WHERE uid = ?", uid).
		Scan(&head)
	if err != nil {
		log.Printf("getUserHead query failed:%d %v", uid, err)
	}
	return head
}

func getSuffix(headurl string) string {
	suffix := ".jpg"
	pos := strings.LastIndex(headurl, ".")
	if pos != -1 {
		suffix = headurl[pos:]
	}
	return suffix
}

func doFetch(msg *nsq.Message) {
	uid := extractUid(msg)
	if uid == 0 {
		log.Printf("doFetch get uid failed:%s", string(msg.Body))
		return
	}
	headurl := getUserHead(db, uid)
	if headurl == "" {
		return
	}
	buf, err := util.DownFile(headurl)
	if err != nil {
		log.Printf("doFetch DownFile failed:%s %v", headurl, err)
		return
	}
	suffix := getSuffix(headurl)
	filename := util.GenUUID() + suffix
	if !ucloud.PutFile(ucloud.Bucket, filename, buf) {
		log.Printf("doFetch ucloud PutFile failed:%d %s", uid, headurl)
		return
	}
	_, err = db.Exec("UPDATE users SET headurl = ? WHERE uid = ?", filename, uid)
	if err != nil {
		log.Printf("doFetch update user headurl failed:%d %s %v",
			uid, filename, err)
	}
}

func fetch(logChan chan *nsq.Message) {
	for {
		select {
		case msg := <-logChan:
			doFetch(msg)
		}
	}
}

var db *sql.DB

func main() {
	done := make(chan bool)
	var err error
	db, err = util.InitDB(false)
	if err != nil {
		log.Fatal(err)
	}
	config := nsq.NewConfig()
	laddr := "127.0.0.1"
	config.LocalAddr, _ = net.ResolveTCPAddr("tcp", laddr+":0")
	config.DefaultRequeueDelay = 0
	config.MaxBackoffDuration = time.Millisecond * 50

	q, err := nsq.NewConsumer("register_head", "ch", config)
	if err != nil {
		log.Fatal(err)
	}
	logChan := make(chan *nsq.Message, 100)
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		logChan <- m
		return nil
	}))

	err = q.ConnectToNSQLookupd("127.0.0.1:4161")
	if err != nil {
		log.Printf("connect failed:%v", err)
	}
	go fetch(logChan)
	<-done
	return
}
