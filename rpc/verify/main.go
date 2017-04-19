package main

import (
	"database/sql"
	"laughing-server/proto/common"
	"laughing-server/proto/verify"
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

func (s *server) FbLogin(ctx context.Context, in *verify.FbLoginRequest) (*verify.LoginReply, error) {
	log.Printf("Login request:%v", in)
	util.PubRPCRequest(w, "verify", "FbLogin")
	uid, token, headurl, nickname, err := fblogin(db, in.Fbid, in.Fbtoken)
	if err != nil {
		return &verify.LoginReply{
			Head: &common.Head{Retcode: common.ErrCode_FB_LOGIN}}, nil
	}
	util.PubRPCSuccRsp(w, "verify", "FbLogin")
	return &verify.LoginReply{
		Head: &common.Head{Retcode: 0},
		Uid:  uid, Token: token, Headurl: headurl, Nickname: nickname}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.VerifyServerPort)
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
	go util.ReportHandler(kv, util.VerifyServerName, util.VerifyServerPort)
	cli := util.InitEtcdCli()
	go util.ReportEtcd(cli, util.VerifyServerName, util.VerifyServerPort)

	s := util.NewGrpcServer()
	verify.RegisterVerifyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
