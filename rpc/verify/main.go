package main

import (
	"database/sql"
	"flag"
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
	uid, token, headurl, err := fblogin(db, in)
	if err != nil {
		return &verify.LoginReply{
			Head: &common.Head{Retcode: common.ErrCode_FB_LOGIN}}, nil
	}
	util.PubRPCSuccRsp(w, "verify", "FbLogin")
	return &verify.LoginReply{
		Head: &common.Head{Retcode: 0},
		Uid:  uid, Token: token, Headurl: headurl}, nil
}

func (s *server) Logout(ctx context.Context, in *common.CommRequest) (*common.CommReply, error) {
	log.Printf("Logout request:%v", in)
	util.PubRPCRequest(w, "verify", "Logout")
	logout(db, in.Head.Uid)
	util.PubRPCSuccRsp(w, "verify", "Logout")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0}}, nil
}

func (s *server) CheckToken(ctx context.Context, in *verify.CheckTokenRequest) (*common.CommReply, error) {
	log.Printf("CheckToken request:%v", in)
	util.PubRPCRequest(w, "verify", "CheckToken")
	flag := checkToken(db, kv, in.Head.Uid, in.Token)
	if !flag {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_TOKEN}}, nil
	}
	util.PubRPCSuccRsp(w, "verify", "CheckToken")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0}}, nil
}

func (s *server) CheckBackToken(ctx context.Context, in *verify.CheckTokenRequest) (*common.CommReply, error) {
	log.Printf("CheckBackToken request:%v", in)
	util.PubRPCRequest(w, "verify", "CheckBackToken")
	flag := checkBackToken(db, in.Head.Uid, in.Token)
	if !flag {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_TOKEN}}, nil
	}
	util.PubRPCSuccRsp(w, "verify", "CheckBackToken")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0}}, nil
}

func (s *server) BackLogin(ctx context.Context, in *verify.BackLoginRequest) (*verify.LoginReply, error) {
	log.Printf("BackLogin request:%v", in)
	util.PubRPCRequest(w, "verify", "BackLogin")
	uid, token, err := backLogin(db, in.Username, in.Passwd)
	if err != nil {
		log.Printf("backLogin failed:%s %s %v", in.Username, in.Passwd, err)
		return &verify.LoginReply{
			Head: &common.Head{Retcode: common.ErrCode_PASSWD}}, nil
	}
	util.PubRPCSuccRsp(w, "verify", "CheckToken")
	return &verify.LoginReply{
		Head: &common.Head{Retcode: 0}, Uid: uid, Token: token}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.VerifyServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	conf := flag.String("conf", util.RpcConfPath, "config file")
	flag.Parse()
	kv, db = util.InitConf(*conf)

	w = util.NewNsqProducer()

	db.SetMaxIdleConns(util.MaxIdleConns)
	go util.ReportHandler(kv, util.VerifyServerName, util.VerifyServerPort)

	s := util.NewGrpcServer()
	verify.RegisterVerifyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
