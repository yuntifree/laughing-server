package main

import (
	"database/sql"
	"flag"
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

func (s *server) Report(ctx context.Context, in *common.CommRequest) (*common.CommReply, error) {
	log.Printf("Report request:%v", in)
	util.PubRPCRequest(w, "modify", "Report")
	err := addReport(db, in.Head.Uid, in.Id)
	if err != nil {
		log.Printf("addReportfailed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_REPORT,
				Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "modify", "Report")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.ModifyServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	conf := flag.String("conf", util.RpcConfPath, "config file")
	flag.Parse()
	kv, db = util.InitConf(*conf)
	w = util.NewNsqProducer()

	db.SetMaxIdleConns(util.MaxIdleConns)
	go util.ReportHandler(kv, util.ModifyServerName, util.ModifyServerPort)

	s := util.NewGrpcServer()
	modify.RegisterModifyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
