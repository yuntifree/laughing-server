package main

import (
	"database/sql"
	"flag"
	"laughing-server/proto/common"
	"laughing-server/proto/limit"
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

func (s *server) CheckDuplicate(ctx context.Context, in *limit.CheckDuplicateRequest) (*common.CommReply, error) {
	log.Printf("CheckDuplicate request:%v", in)
	util.PubRPCRequest(w, "limit", "CheckDuplicate")
	flag := checkDuplicate(kv, in.Type, in.Id)
	if flag {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_LIMIT, Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "limit", "CheckDuplicate")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.LimitServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	conf := flag.String("conf", util.RpcConfPath, "config file")
	flag.Parse()
	kv, db = util.InitConf(*conf)
	w = util.NewNsqProducer()

	db.SetMaxIdleConns(util.MaxIdleConns)
	go util.ReportHandler(kv, util.LimitServerName, util.LimitServerPort)

	s := util.NewGrpcServer()
	limit.RegisterLimitServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
