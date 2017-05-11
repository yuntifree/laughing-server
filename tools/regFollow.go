package main

import (
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

var uids = []int64{
	1,
}

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
	for i := 0; i < len(uids); i++ {
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

func main() {
	done := make(chan bool)
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
