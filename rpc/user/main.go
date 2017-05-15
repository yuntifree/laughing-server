package main

import (
	"database/sql"
	"flag"
	"laughing-server/proto/common"
	"laughing-server/proto/user"
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

func (s *server) GetInfo(ctx context.Context, in *common.CommRequest) (*user.InfoReply, error) {
	log.Printf("GetInfo request:%v", in)
	util.PubRPCRequest(w, "user", "GetInfo")
	info, err := getInfo(db, in.Head.Uid, in.Id)
	if err != nil {
		log.Printf("GetInfo query failed:%v", err)
		return &user.InfoReply{
			Head: &common.Head{Retcode: common.ErrCode_GET_INFO, Uid: in.Head.Uid},
		}, nil
	}
	util.PubRPCSuccRsp(w, "user", "GetInfo")
	return &user.InfoReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Info: &info,
	}, nil
}

func (s *server) FetchInfos(ctx context.Context, in *common.CommRequest) (*user.FetchInfosReply, error) {
	log.Printf("FetchInfos request:%v", in)
	util.PubRPCRequest(w, "user", "FetchInfos")
	infos := fetchInfos(db, in.Seq, in.Num)
	total := getTotalUsers(db)
	util.PubRPCSuccRsp(w, "user", "FetchInfos")
	return &user.FetchInfosReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Total: total,
	}, nil
}

func (s *server) AddInfo(ctx context.Context, in *user.InfoRequest) (*common.CommReply, error) {
	log.Printf("AddInfo request:%v", in)
	util.PubRPCRequest(w, "user", "AddInfo")
	id, err := addInfo(db, in.Info)
	util.PubRPCSuccRsp(w, "user", "AddInfo")
	if err != nil {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_ADD_INFO, Uid: in.Head.Uid},
		}, nil
	}
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Id:   id,
	}, nil
}

func (s *server) ModInfo(ctx context.Context, in *user.ModInfoRequest) (*common.CommReply, error) {
	log.Printf("ModInfo request:%v", in)
	util.PubRPCRequest(w, "user", "ModInfo")
	err := modInfo(db, in.Info)
	if err != nil {
		log.Printf("modInfo failed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_MOD_INFO,
				Uid: in.Head.Uid},
		}, nil
	}
	util.PubRPCSuccRsp(w, "user", "ModInfo")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.UserServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	conf := flag.String("conf", util.RpcConfPath, "config file")
	flag.Parse()

	w = util.NewNsqProducer()
	kv, db = util.InitConf(*conf)

	db.SetMaxIdleConns(util.MaxIdleConns)
	go util.ReportHandler(kv, util.UserServerName, util.UserServerPort)

	s := util.NewGrpcServer()
	user.RegisterUserServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
