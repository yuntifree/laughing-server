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

const (
	emptyRespCode = 999
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

func (s *server) ModVersion(ctx context.Context, in *config.VersionRequest) (*common.CommReply, error) {
	log.Printf("ModVersion request:%v", in)
	util.PubRPCRequest(w, "config", "ModVersion")
	err := modVersion(db, in.Info)
	if err != nil {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_MOD_VERSION, Uid: in.Head.Uid},
		}, nil
	}
	util.PubRPCSuccRsp(w, "config", "ModVersion")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func (s *server) FetchUserLang(ctx context.Context, in *common.CommRequest) (*config.LangReply, error) {
	log.Printf("FetchUserLang request:%v", in)
	util.PubRPCRequest(w, "config", "FetchUserLang")
	infos := fetchUserLang(db)
	if len(infos) == 0 {
		return &config.LangReply{
			Head: &common.Head{Retcode: emptyRespCode, Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "config", "FetchUserLang")
	return &config.LangReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}, Infos: infos}, nil
}

func (s *server) AddUserLang(ctx context.Context, in *config.LangRequest) (*common.CommReply, error) {
	log.Printf("AddUserLang request:%v", in)
	util.PubRPCRequest(w, "config", "AddUserLang")
	id, err := addUserLang(db, in.Info)
	if err != nil {
		log.Printf("AddUserLang failed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: 1, Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "config", "AddUserLang")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}, Id: id}, nil
}

func (s *server) DelUserLang(ctx context.Context, in *common.CommRequest) (*common.CommReply, error) {
	log.Printf("DelUserLang request:%v", in)
	util.PubRPCRequest(w, "config", "DelUserLang")
	err := delUserLang(db, in.Id)
	if err != nil {
		log.Printf("DelUserLang failed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: 1, Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "config", "DelUserLang")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func (s *server) FetchLangFollow(ctx context.Context, in *common.CommRequest) (*config.LangFollowReply, error) {
	log.Printf("FetchLangFollow request:%v", in)
	util.PubRPCRequest(w, "config", "FetchLangFollow")
	infos := fetchLangFollow(db)
	if len(infos) == 0 {
		return &config.LangFollowReply{
			Head: &common.Head{Retcode: emptyRespCode, Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "config", "FetchLangFollow")
	return &config.LangFollowReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}, Infos: infos}, nil
}

func (s *server) AddLangFollow(ctx context.Context, in *config.LangFollowRequest) (*common.CommReply, error) {
	log.Printf("AddLangFollow request:%v", in)
	util.PubRPCRequest(w, "config", "AddLangFollow")
	addLangFollow(db, in.Lid, in.Uids)
	util.PubRPCSuccRsp(w, "config", "AddLangFollow")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func (s *server) DelLangFollow(ctx context.Context, in *config.DelLangFollowRequest) (*common.CommReply, error) {
	log.Printf("DelLangFollow request:%v", in)
	util.PubRPCRequest(w, "config", "DelLangFollow")
	err := delLangFollow(db, in.Ids)
	if err != nil {
		log.Printf("DelLangFollow failed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: 1, Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "config", "DelLangFollow")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
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
