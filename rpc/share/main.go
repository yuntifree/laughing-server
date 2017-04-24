package main

import (
	"database/sql"
	"laughing-server/proto/common"
	"laughing-server/proto/share"
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

func (s *server) GetTags(ctx context.Context, in *common.CommRequest) (*share.TagReply, error) {
	log.Printf("GetTags request:%v", in)
	util.PubRPCRequest(w, "share", "GetTags")
	infos := getTags(db)
	util.PubRPCSuccRsp(w, "share", "GetTags")
	return &share.TagReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos}, nil
}

func (s *server) AddShare(ctx context.Context, in *share.ShareRequest) (*common.CommReply, error) {
	log.Printf("AddShare request:%v", in)
	util.PubRPCRequest(w, "share", "AddShare")
	id, err := addShare(db, in)
	if err != nil {
		log.Printf("addShare failed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_ADD_SHARE,
				Uid: in.Head.Uid}}, nil

	}
	util.PubRPCSuccRsp(w, "share", "AddShare")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Id:   id}, nil
}

func (s *server) Reshare(ctx context.Context, in *common.CommRequest) (*common.CommReply, error) {
	log.Printf("Reshare request:%v", in)
	util.PubRPCRequest(w, "share", "Reshare")
	id, err := reshare(db, in.Head.Uid, in.Id)
	if err != nil {
		log.Printf("reshare failed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_RESHARE,
				Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "share", "Reshare")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Id:   id}, nil
}

func (s *server) AddComment(ctx context.Context, in *share.CommentRequest) (*common.CommReply, error) {
	log.Printf("AddComment request:%v", in)
	util.PubRPCRequest(w, "share", "AddComment")
	id, err := addComment(db, in.Head.Uid, in.Id, in.Content)
	if err != nil {
		log.Printf("addComment failed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_ADD_COMMENT,
				Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "share", "AddComment")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Id:   id}, nil
}

func (s *server) GetMyShares(ctx context.Context, in *common.CommRequest) (*share.ShareReply, error) {
	log.Printf("GetMyShares request:%v", in)
	util.PubRPCRequest(w, "share", "GetMyShare")
	infos, nextseq := getMyShares(db, in.Head.Uid, in.Seq, in.Num)
	var hasmore int64
	if len(infos) >= int(in.Num) {
		hasmore = 1
	}
	util.PubRPCSuccRsp(w, "share", "GetMyShare")
	return &share.ShareReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Hasmore: hasmore, Nextseq: nextseq}, nil
}

func (s *server) GetShares(ctx context.Context, in *common.CommRequest) (*share.ShareReply, error) {
	log.Printf("GetShares request:%v", in)
	util.PubRPCRequest(w, "share", "GetShare")
	infos, nextseq := getShares(db, in.Seq, in.Num)
	var hasmore int64
	if len(infos) >= int(in.Num) {
		hasmore = 1
	}
	util.PubRPCSuccRsp(w, "share", "GetShare")
	return &share.ShareReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Hasmore: hasmore, Nextseq: nextseq}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.ShareServerPort)
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
	go util.ReportHandler(kv, util.ShareServerName, util.ShareServerPort)
	cli := util.InitEtcdCli()
	go util.ReportEtcd(cli, util.ShareServerName, util.ShareServerPort)

	s := util.NewGrpcServer()
	share.RegisterShareServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
