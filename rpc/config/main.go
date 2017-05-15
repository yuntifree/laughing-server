package main

import (
	"database/sql"
	"flag"
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

func (s *server) FetchVersions(ctx context.Context, in *common.CommRequest) (*config.FetchVersionReply, error) {
	log.Printf("FetchVersion request:%v", in)
	util.PubRPCRequest(w, "config", "FetchVersions")
	infos := fetchVersions(db, in.Seq, in.Num)
	total := getTotalVersions(db)
	util.PubRPCSuccRsp(w, "config", "FetchVersions")
	return &config.FetchVersionReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Total: total}, nil
}

func (s *server) AddVersion(ctx context.Context, in *config.VersionRequest) (*common.CommReply, error) {
	log.Printf("AddVersion request:%v", in)
	util.PubRPCRequest(w, "config", "AddVersion")
	id, err := addVersion(db, in.Info)
	if err != nil {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_ADD_VERSION, Uid: in.Head.Uid},
		}, nil
	}
	util.PubRPCSuccRsp(w, "config", "AddVersion")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Id:   id}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.ConfigServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	conf := flag.String("conf", util.RpcConfPath, "config file")
	flag.Parse()
	kv, db = util.InitConf(*conf)
	w = util.NewNsqProducer()

	db.SetMaxIdleConns(util.MaxIdleConns)
	go util.ReportHandler(kv, util.ConfigServerName, util.ConfigServerPort)

	s := util.NewGrpcServer()
	config.RegisterConfigServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
