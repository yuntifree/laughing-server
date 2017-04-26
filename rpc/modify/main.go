package main

import (
	"database/sql"
	"laughing-server/proto/common"
	"laughing-server/proto/modify"
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

func (s *server) ReportClick(ctx context.Context, in *modify.ClickRequest) (*common.CommReply, error) {
	log.Printf("ReportClick request:%v", in)
	util.PubRPCRequest(w, "modify", "ReportClick")
	err := reportClick(db, in)
	if err != nil {
		log.Printf("reportClick failed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_REPORT_CLICK,
				Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "modify", "ReportClick")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.ModifyServerPort)
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
	go util.ReportHandler(kv, util.ModifyServerName, util.ModifyServerPort)
	cli := util.InitEtcdCli()
	go util.ReportEtcd(cli, util.ModifyServerName, util.ModifyServerPort)

	s := util.NewGrpcServer()
	modify.RegisterModifyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
