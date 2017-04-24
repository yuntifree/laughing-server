package main

import (
	"laughing-server/httpserver"
	"laughing-server/proto/common"
	"laughing-server/proto/fan"
	"laughing-server/proto/share"
	"laughing-server/proto/user"
	"laughing-server/proto/verify"
	"laughing-server/util"
	"log"
	"net/http"
)

func followop(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	uid := req.GetParamInt("uid")
	tuid := req.GetParamInt("tuid")
	ftype := req.GetParamInt("type")
	log.Printf("followop uid:%d tuid:%d type:%d", uid, tuid, ftype)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.FanServerType, uid, "Follow",
		&fan.FanRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Type: ftype, Tuid: tuid})

	httpserver.CheckRPCErr(rpcerr, "Follow")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "Follow")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func fblogin(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	fbid := req.GetParamString("fb_id")
	fbtoken := req.GetParamString("fb_token")
	log.Printf("fblogin fb_id:%s fb_token:%s", fbid, fbtoken)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.VerifyServerType, 0, "FbLogin",
		&verify.FbLoginRequest{Head: &common.Head{Sid: uuid},
			Fbid: fbid, Fbtoken: fbtoken})

	httpserver.CheckRPCErr(rpcerr, "FbLogin")
	res := resp.Interface().(*verify.LoginReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FbLogin")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func logout(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	uid := req.GetParamInt("uid")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.VerifyServerType, uid, "Logout",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: uid}})

	httpserver.CheckRPCErr(rpcerr, "Logout")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "Logout")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getRelations(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	uid := req.GetParamInt("uid")
	stype := req.GetParamInt("type")
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.FanServerType, uid, "GetRelations",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Type: stype, Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "GetRelations")
	res := resp.Interface().(*fan.RelationReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetRelations")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getTags(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, 0, "GetTags",
		&common.CommRequest{Head: &common.Head{Sid: uuid}})

	httpserver.CheckRPCErr(rpcerr, "GetTags")
	res := resp.Interface().(*share.TagReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetTags")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addShare(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	uid := req.GetParamInt("uid")
	title := req.GetParamString("title")
	abstract := req.GetParamString("abstract")
	img := req.GetParamString("img")
	dst := req.GetParamString("dst")
	origin := req.GetParamInt("origin")
	tags := req.GetParamIntArray("tags")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, 0, "AddShare",
		&share.ShareRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Title: title, Abstract: abstract, Img: img,
			Dst: dst, Tags: tags, Origin: origin})

	httpserver.CheckRPCErr(rpcerr, "AddShare")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddShare")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func reshare(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	uid := req.GetParamInt("uid")
	id := req.GetParamInt("id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, uid, "Reshare",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Id: id})

	httpserver.CheckRPCErr(rpcerr, "Reshare")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "Reshare")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addComment(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	uid := req.GetParamInt("uid")
	id := req.GetParamInt("id")
	content := req.GetParamString("content")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, uid, "AddComment",
		&share.CommentRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Id: id, Content: content})

	httpserver.CheckRPCErr(rpcerr, "AddComment")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddComment")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getMyShares(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	uid := req.GetParamInt("uid")
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, uid, "GetMyShares",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "GetMyShares")
	res := resp.Interface().(*share.ShareReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetMyShares")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getShares(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	uid := req.GetParamInt("uid")
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, uid, "GetShares",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "GetShares")
	res := resp.Interface().(*share.ShareReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetShares")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getUserInfo(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	uid := req.GetParamInt("uid")
	tuid := req.GetParamInt("tuid")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.UserServerType, uid, "GetInfo",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: uid}, Id: tuid})

	httpserver.CheckRPCErr(rpcerr, "GetInfo")
	res := resp.Interface().(*user.InfoReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetInfo")

	body := httpserver.GenInfoResponseBody(res)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

//NewAppServer return app http handler
func NewAppServer() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/followop", httpserver.AppHandler(followop))
	mux.Handle("/fblogin", httpserver.AppHandler(fblogin))
	mux.Handle("/logout", httpserver.AppHandler(logout))
	mux.Handle("/get_relations", httpserver.AppHandler(getRelations))
	mux.Handle("/get_tags", httpserver.AppHandler(getTags))
	mux.Handle("/add_share", httpserver.AppHandler(addShare))
	mux.Handle("/get_my_shares", httpserver.AppHandler(getMyShares))
	mux.Handle("/get_shares", httpserver.AppHandler(getShares))
	mux.Handle("/reshare", httpserver.AppHandler(reshare))
	mux.Handle("/add_comment", httpserver.AppHandler(addComment))
	mux.Handle("/get_user_info", httpserver.AppHandler(getUserInfo))
	return mux
}
