package main

import (
	"database/sql"
	"laughing-server/proto/common"
	"laughing-server/proto/config"
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

func (s *server) CheckUpdate(ctx context.Context, in *common.CommRequest) (*config.VersionReply, error) {
	log.Printf("CheckUpdate request:%v", in)
	util.PubRPCRequest(w, "config", "CheckUpdate")
	version, desc, title, subtitle, downurl := checkUpdate(db, in.Head.Term, in.Head.Version)
	if version == "" {
		return &config.VersionReply{
			Head: &common.Head{Retcode: common.ErrCode_CHECK_UPDATE,
				Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "config", "CheckUpate")
	return &config.VersionReply{
		Head:    &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Version: version, Desc: desc, Title: title, Subtitle: subtitle,
		Downurl: downurl}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.ConfigServerPort)
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
	go util.ReportHandler(kv, util.ConfigServerName, util.ConfigServerPort)
	cli := util.InitEtcdCli()
	go util.ReportEtcd(cli, util.ConfigServerName, util.ConfigServerPort)

	s := util.NewGrpcServer()
	config.RegisterConfigServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
