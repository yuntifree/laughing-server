package main

import (
	"database/sql"
	"flag"
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
	log.Printf("Follow request:%v", in)
	util.PubRPCRequest(w, "fan", "Follow")
	if in.Head.Uid == in.Tuid {
		log.Printf("Follow illegal same uid:%d", in.Head.Uid)
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_FOLLOW, Uid: in.Head.Uid}}, nil
	}
	flag := follow(db, in.Type, in.Head.Uid, in.Tuid)
	if !flag {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_FOLLOW, Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "fan", "Follow")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func (s *server) GetRelations(ctx context.Context, in *common.CommRequest) (*fan.RelationReply, error) {
	log.Printf("GetRelations request:%v", in)
	util.PubRPCRequest(w, "fan", "GetRelations")
	infos, nextseq := getRelations(db, in.Head.Uid, in.Id, in.Type, in.Seq, in.Num)
	var hasmore int64
	if len(infos) >= int(in.Num) {
		hasmore = 1
	}
	util.PubRPCSuccRsp(w, "fan", "GetRelations")
	return &fan.RelationReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Hasmore: hasmore, Nextseq: nextseq}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.FanServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	conf := flag.String("conf", util.RpcConfPath, "config file")
	flag.Parse()
	kv, db = util.InitConf(*conf)
	w = util.NewNsqProducer()

	db.SetMaxIdleConns(util.MaxIdleConns)
	go util.ReportHandler(kv, util.FanServerName, util.FanServerPort)

	s := util.NewGrpcServer()
	fan.RegisterFanServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
