package main

import (
	"database/sql"
	"laughing-server/httpserver"
	"laughing-server/proto/common"
	"laughing-server/proto/fan"
	"laughing-server/util"
	"log"
	"net"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	_ "github.com/go-sql-driver/mysql"
	nsq "github.com/nsqio/go-nsq"
)

func follow(uid, tuid int64) {
	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.FanServerType, uid, "Follow",
		&fan.FanRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Type: 0, Tuid: tuid})
	if rpcerr.Interface() != nil {
		log.Printf("follow rpc failed:%d %d %v", uid, tuid, rpcerr.Interface())
		return
	}

	res := resp.Interface().(*common.CommReply)
	if res.Head.Retcode != 0 {
		log.Printf("follow failed:%d %d ret:%d", uid, tuid, res.Head.Retcode)
	}
	return
}

func getLangUids(lang string) []int64 {
	if v, ok := langUids[lang]; ok {
		return v
	}
	return recommendUids
}

func doFollow(msg *nsq.Message) {
	log.Printf("doFollow get msg:%s", string(msg.Body))
	js, err := simplejson.NewJson(msg.Body)
	if err != nil {
		log.Printf("HandlerMessage parse body failed:%s %v",
			string(msg.Body), err)
		return
	}
	uid, err := js.Get("uid").Int64()
	if err != nil {
		log.Printf("get uid failed:%s %v", string(msg.Body), err)
		return
	}
	lang, _ := js.Get("lang").String()
	if time.Now().After(loadTime.Add(1 * time.Hour)) {
		loadLangUids()
		loadTime = time.Now()
	}
	uids := getLangUids(lang)
	log.Printf("lang:%s uids:%+v", lang, uids)
	for i := 0; i < len(uids); i++ {
		log.Printf("follow uid:%d tuid:%d", uid, uids[i])
		follow(uid, uids[i])
	}
}

func followUsers(logChan chan *nsq.Message) {
	for {
		select {
		case msg := <-logChan:
			doFollow(msg)
		}
	}
}

func getRecommendUids(db *sql.DB) ([]int64, error) {
	var uids []int64
	rows, err := db.Query("SELECT uid FROM users WHERE recommend = 1")
	if err != nil {
		return uids, err
	}
	defer rows.Close()
	for rows.Next() {
		var uid int64
		err = rows.Scan(&uid)
		if err != nil {
			continue
		}
		uids = append(uids, uid)
	}
	return uids, nil
}

func loadLangUids() {
	db, err := util.InitDB(false)
	if err != nil {
		log.Printf("loadLangUids InitDB failed:%v", err)
		return
	}
	defer db.Close()
	recommendUids, err := getRecommendUids(db)
	if err != nil {
		log.Printf("loadLangUids getRecommendUids failed:%v", err)
	}
	log.Printf("recommendUids:%+v", recommendUids)
	rows, err := db.Query("SELECT l.lang, f.uid FROM user_lang l, lang_follower f WHERE l.id = f.lid AND f.deleted = 0 AND l.deleted = 0 ORDER BY l.lang")
	if err != nil {
		log.Printf("loadLangUids query failed:%v", err)
		return
	}
	defer rows.Close()
	var def string
	var uids []int64
	for rows.Next() {
		var lang string
		var uid int64
		err = rows.Scan(&lang, &uid)
		if err != nil {
			log.Printf("loadLangUids scan failed:%v", err)
			continue
		}
		if def != "" && def != lang {
			if len(uids) > 0 {
				log.Printf("def:%s uids:%v", def, uids)
				langUids[def] = uids
			}
			def = lang
			uids = uids[0:0]
			uids = append(uids, uid)
		} else {
			if def == "" {
				def = lang
			}
			uids = append(uids, uid)
		}
	}
	if len(uids) > 0 {
		langUids[def] = uids
	}
}

var recommendUids []int64
var loadTime time.Time
var langUids map[string][]int64

func main() {
	done := make(chan bool)
	langUids = make(map[string][]int64)
	var err error
	config := nsq.NewConfig()
	laddr := "127.0.0.1"
	config.LocalAddr, _ = net.ResolveTCPAddr("tcp", laddr+":0")
	config.DefaultRequeueDelay = 0
	config.MaxBackoffDuration = time.Millisecond * 50

	q, err := nsq.NewConsumer("register", "ch", config)
	if err != nil {
		log.Fatal(err)
	}
	loadLangUids()
	loadTime = time.Now()
	logChan := make(chan *nsq.Message, 100)
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		logChan <- m
		return nil
	}))

	err = q.ConnectToNSQLookupd("127.0.0.1:4161")
	if err != nil {
		log.Printf("connect failed:%v", err)
	}
	go followUsers(logChan)
	<-done
	return
}
