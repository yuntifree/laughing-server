package main

import (
	"database/sql"
	"flag"
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

const (
	recommendTag = 1
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

func (s *server) FetchTags(ctx context.Context, in *common.CommRequest) (*share.TagReply, error) {
	log.Printf("FetchTags request:%v", in)
	util.PubRPCRequest(w, "share", "FetchTags")
	infos := fetchTags(db, in.Seq, in.Num)
	total := getTotalTags(db)
	util.PubRPCSuccRsp(w, "share", "FetchTags")
	return &share.TagReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Total: total}, nil
}

func (s *server) AddTag(ctx context.Context, in *share.TagRequest) (*common.CommReply, error) {
	log.Printf("AddTag request:%v", in)
	util.PubRPCRequest(w, "share", "AddTag")
	id, err := addTag(db, in.Info)
	if err != nil {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_ADD_TAG}}, nil
	}
	util.PubRPCSuccRsp(w, "share", "AddTag")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Id:   id}, nil
}

func (s *server) ModTag(ctx context.Context, in *share.TagRequest) (*common.CommReply, error) {
	log.Printf("ModTag request:%v", in)
	util.PubRPCRequest(w, "share", "ModTag")
	err := modTag(db, in.Info)
	if err != nil {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_ADD_TAG}}, nil
	}
	util.PubRPCSuccRsp(w, "share", "ModTag")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func (s *server) DelTags(ctx context.Context, in *share.DelTagRequest) (*common.CommReply, error) {
	log.Printf("DelTag request:%v", in)
	util.PubRPCRequest(w, "share", "DelTags")
	err := delTags(db, in.Ids)
	if err != nil {
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_DEL_TAG}}, nil
	}
	util.PubRPCSuccRsp(w, "share", "DelTags")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
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

func (s *server) Unshare(ctx context.Context, in *common.CommRequest) (*common.CommReply, error) {
	log.Printf("Unshare request:%v", in)
	util.PubRPCRequest(w, "share", "Unshare")
	err := unshare(db, in.Head.Uid, in.Id)
	if err != nil {
		log.Printf("unshare failed:%v", err)
		return &common.CommReply{
			Head: &common.Head{Retcode: common.ErrCode_UNSHARE,
				Uid: in.Head.Uid}}, nil
	}
	util.PubRPCSuccRsp(w, "share", "Unshare")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
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

func getHasmore(len int, num int64) int64 {
	if len >= int(num) {
		return 1
	}
	return 0
}

func (s *server) GetUserShares(ctx context.Context, in *common.CommRequest) (*share.ShareReply, error) {
	log.Printf("GetUserShares request:%v", in)
	util.PubRPCRequest(w, "share", "GetUserShare")
	infos, nextseq := getUserShares(db, in.Head.Uid, in.Id, in.Seq, in.Num)
	hasmore := getHasmore(len(infos), in.Num)
	util.PubRPCSuccRsp(w, "share", "GetUserShare")
	return &share.ShareReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Hasmore: hasmore, Nextseq: nextseq}, nil
}

func (s *server) GetShares(ctx context.Context, in *common.CommRequest) (*share.ShareReply, error) {
	log.Printf("GetShares request:%v", in)
	util.PubRPCRequest(w, "share", "GetShare")
	infos, nextseq := getShares(db, in.Head.Uid, in.Seq, in.Num, in.Id)
	hasmore := getHasmore(len(infos), in.Num)
	recommend := getRecommendTag(db)
	util.PubRPCSuccRsp(w, "share", "GetShare")
	return &share.ShareReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Hasmore: hasmore, Nextseq: nextseq,
		Recommendtag: recommend}, nil
}

func (s *server) FetchShares(ctx context.Context, in *common.CommRequest) (*share.ShareReply, error) {
	log.Printf("FetchShares request:%v", in)
	util.PubRPCRequest(w, "share", "FetchShare")
	infos := fetchShares(db, in.Seq, in.Num, in.Type)
	total := getTotalShares(db, in.Type)
	util.PubRPCSuccRsp(w, "share", "GetShare")
	return &share.ShareReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Total: total}, nil
}

func (s *server) SearchShare(ctx context.Context, in *common.CommRequest) (*share.ShareReply, error) {
	log.Printf("SearchShares request:%v", in)
	util.PubRPCRequest(w, "share", "SearchShare")
	infos := searchShares(db, in.Id)
	util.PubRPCSuccRsp(w, "share", "SearchShare")
	return &share.ShareReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos}, nil
}

func (s *server) GetShareIds(ctx context.Context, in *common.CommRequest) (*share.ShareIdReply, error) {
	log.Printf("GetShareIds request:%v", in)
	util.PubRPCRequest(w, "share", "GetShareIds")
	ids, nextseq, nexttag := getShareIds(db, in.Seq, in.Num, in.Id, 0)
	hasmore := getHasmore(len(ids), in.Num)
	util.PubRPCSuccRsp(w, "share", "GetShareIds")
	return &share.ShareIdReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Ids:  ids, Hasmore: hasmore, Nextseq: nextseq,
		Nexttag: nexttag}, nil
}

func (s *server) GetShareComments(ctx context.Context, in *common.CommRequest) (*share.CommentReply, error) {
	log.Printf("GetShareComments request:%v", in)
	util.PubRPCRequest(w, "share", "GetShareComments")
	infos, nextseq := getShareComments(db, in.Id, in.Seq, in.Num)
	hasmore := getHasmore(len(infos), in.Num)
	util.PubRPCSuccRsp(w, "share", "GetShareComments")
	return &share.CommentReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos, Hasmore: hasmore, Nextseq: nextseq}, nil
}

func (s *server) GetShareDetail(ctx context.Context, in *common.CommRequest) (*share.ShareDetailReply, error) {
	log.Printf("GetShareDetail request:%v", in)
	util.PubRPCRequest(w, "share", "GetShareDetail")
	info, err := getShareDetail(db, in.Head.Uid, in.Id)
	if err != nil {
		log.Printf("getShareDetail failed:%v", err)
		return &share.ShareDetailReply{
			Head: &common.Head{Retcode: common.ErrCode_GET_INFO, Uid: in.Head.Uid},
		}, nil
	}
	util.PubRPCSuccRsp(w, "share", "GetShareDetail")
	return &share.ShareDetailReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Info: &info}, nil
}

func (s *server) GetRecommendShares(ctx context.Context, in *share.RecommendShareRequest) (*share.RecommendShareReply, error) {
	log.Printf("GetRecommendSharerequest:%v", in)
	util.PubRPCRequest(w, "share", "GetRecommendShares")
	infos, err := getRecommendShares(db, in.Head.Uid, in.Tagid, in.Sid)
	if err != nil {
		log.Printf("getRecommendSharefailed:%v", err)
		return &share.RecommendShareReply{
			Head: &common.Head{Retcode: common.ErrCode_GET_INFO, Uid: in.Head.Uid},
		}, nil
	}
	util.PubRPCSuccRsp(w, "share", "GetRecommendShare")
	return &share.RecommendShareReply{
		Head:  &common.Head{Retcode: 0, Uid: in.Head.Uid},
		Infos: infos}, nil
}

func (s *server) ReviewShare(ctx context.Context, in *share.ReviewShareRequest) (*common.CommReply, error) {
	log.Printf("ReviewShare request:%v", in)
	util.PubRPCRequest(w, "share", "ReviewShare")
	reviewShare(db, in)
	util.PubRPCSuccRsp(w, "share", "ReviewShare")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func (s *server) AddShareTags(ctx context.Context, in *share.AddTagRequest) (*common.CommReply, error) {
	log.Printf("AddShareTags request:%v", in)
	util.PubRPCRequest(w, "share", "AddShareTags")
	addShareTag(db, in.Id, in.Tags)
	util.PubRPCSuccRsp(w, "share", "AddShareTags")
	return &common.CommReply{
		Head: &common.Head{Retcode: 0, Uid: in.Head.Uid}}, nil
}

func main() {
	lis, err := net.Listen("tcp", util.ShareServerPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	conf := flag.String("conf", util.RpcConfPath, "config file")
	flag.Parse()
	kv, db = util.InitConf(*conf)
	w = util.NewNsqProducer()

	db.SetMaxIdleConns(util.MaxIdleConns)
	go util.ReportHandler(kv, util.ShareServerName, util.ShareServerPort)

	s := util.NewGrpcServer()
	share.RegisterShareServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
