package main

import (
	"database/sql"
	"laughing-server/proto/common"
	"laughing-server/proto/fan"
	"laughing-server/util"
	"log"
	"net"

	"golang.org/x/net/context"

	_ "github.com/go-sql-driver/mysql"
	nsq "github.com/nsqio/go-nsq"
	redis "gopkg.in/redis.v5"
)

type server struct{}

var db *sql.DB
var kv *redis.Client
var w *nsq.Producer

func (s *server) Follow(ctx context.Context, in *fan.FanRequest) (*common.CommReply, error) {
	util.PubRPCRequest(w, "fan", "Follow")
	follow(db, in.Head.Uid, in.Type, in.Tuid)
	util.PubRPCSuccRsp(w, "fan", "Follow")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.FanServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	w = util.NewNsqProducer()

	db, err = util.InitDB(false)
	if err != nil {
		log.Fatalf("failed to init db connection: %v", err)
	}
	db.SetMaxIdleConns(util.MaxIdleConns)
	kv = util.InitRedis()
	go util.ReportHandler(kv, util.FanServerName, util.FanServerPort)
	cli := util.InitEtcdCli()
	go util.ReportEtcd(cli, util.FanServerName, util.FanServerPort)

	s := util.NewGrpcServer()
	fan.RegisterFanServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
