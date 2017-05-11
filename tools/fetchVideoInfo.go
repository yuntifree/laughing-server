package main

import (
	"database/sql"
	"fmt"
	"laughing-server/spider"
	"laughing-server/ucloud"
	"laughing-server/util"
	"log"
	"net"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	nsq "github.com/nsqio/go-nsq"
)

const (
	facebook  = 1
	instagram = 2
	musically = 3
	youtube   = 4
)

func getVideoInfo(origin int64, dst string) *spider.VideoInfo {
	var info *spider.VideoInfo
	switch origin {
	case facebook:
		info = spider.GetFacebook(dst)
	case instagram:
		info = spider.GetInstagram(dst)
	case musically:
		info = spider.GetMusically(dst)
	case youtube:
		info = spider.GetYoutube(dst)
	}
	return info
}

func extractSid(msg *nsq.Message) int64 {
	js, err := simplejson.NewJson(msg.Body)
	if err != nil {
		log.Printf("HandlerMessage parse body failed:%s %v",
			string(msg.Body), err)
		return 0
	}
	sid, err := js.Get("sid").Int64()
	if err != nil {
		log.Printf("get uid failed:%s %v", string(msg.Body), err)
		return 0
	}
	return sid
}

func handleFile(url string) (string, error) {
	buf, err := util.DownFile(url)
	if err != nil {
		return "", err
	}
	filename := util.GenUUID() + ".jpg"
	if !ucloud.PutFile(ucloud.Bucket, filename, buf) {
		log.Printf("handleImg ucloud PutFile failed:%s", filename)
		return "", fmt.Errorf("ucloud PutFile failed:%s", filename)
	}
	return filename, nil
}

func handleImg(url string) (string, error) {
	return handleFile(url)
}

func handleVideo(url string) (string, error) {
	return handleFile(url)
}

func doFetch(msg *nsq.Message) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic :%v", r)
		}
	}()
	log.Printf("msg:%s", string(msg.Body))
	sid := extractSid(msg)
	if sid == 0 {
		return
	}
	var dst string
	var mid, origin, done int64
	err := db.QueryRow("SELECT m.id, m.dst, m.origin, m.done FROM media m, shares s WHERE s.mid = m.id AND s.id = ?", sid).
		Scan(&mid, &dst, &origin, &done)
	if err != nil {
		log.Printf("doFetch scan failed:%d %v", sid, err)
		return
	}
	if done > 0 {
		log.Printf("sid:%d has done", sid)
		return
	}
	info := getVideoInfo(origin, dst)
	if info == nil {
		log.Printf("get video info failed:%d %s", origin, dst)
		return
	}
	log.Printf("sid:%d video info:%v", sid, info)
	img, err := handleImg(info.ThumbUrl)
	if err != nil {
		log.Printf("handleImg failed:%v", err)
		return
	}
	video, err := handleVideo(info.VideoUrl)
	if err != nil {
		log.Printf("handleVideo failed:%v", err)
		return
	}
	_, err = db.Exec("UPDATE media SET width = ?, height = ?, cdn = ?, img = ?, src = ?, done = 1 WHERE id = ?",
		info.Width, info.Height, video, img, dst, mid)
	if err != nil {
		log.Printf("update media info failed sid:%d video info:%v, img:%s video:%s",
			sid, info, img, video)
	}
	return
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

	q, err := nsq.NewConsumer("shares", "ch", config)
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
