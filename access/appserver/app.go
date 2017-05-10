package main

import (
	"laughing-server/httpserver"
	"laughing-server/proto/common"
	"laughing-server/proto/config"
	"laughing-server/proto/fan"
	"laughing-server/proto/modify"
	"laughing-server/proto/share"
	"laughing-server/proto/user"
	"laughing-server/proto/verify"
	"laughing-server/util"
	"log"
	"net/http"
)

func followop(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitCheck(r)
	tuid := req.GetParamInt("tuid")
	ftype := req.GetParamInt("type")
	log.Printf("followop uid:%d tuid:%d type:%d", req.Uid, tuid, ftype)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.FanServerType, req.Uid, "Follow",
		&fan.FanRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
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
	nickname := req.GetParamString("nickname")
	dev := req.ParseDevice(r)
	log.Printf("fblogin fb_id:%s fb_token:%s nickname:%s device:%v", fbid, fbtoken, nickname, dev)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.VerifyServerType, 0, "FbLogin",
		&verify.FbLoginRequest{Head: &common.Head{Sid: uuid},
			Fbid: fbid, Fbtoken: fbtoken, Imei: dev.Imei, Model: dev.Model,
			Language: dev.Language, Version: dev.Version, Os: dev.Os,
			Nickname: nickname})

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

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.VerifyServerType, req.Uid, "Logout",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid}})

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
	stype := req.GetParamInt("type")
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")
	tuid := req.GetParamInt("tuid")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.FanServerType, req.Uid, "GetRelations",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Type: stype, Seq: seq, Num: num, Id: tuid})

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

	title := req.GetParamStringDef("title", "")
	img := req.GetParamStringDef("img", "")
	dst := req.GetParamString("dst")
	origin := req.GetParamInt("origin")
	tags := req.GetParamIntArray("tags")
	width := req.GetParamIntDef("width", 0)
	height := req.GetParamIntDef("height", 0)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "AddShare",
		&share.ShareRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Title: title, Img: img,
			Dst: dst, Tags: tags, Origin: origin,
			Width: width, Height: height})

	httpserver.CheckRPCErr(rpcerr, "AddShare")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddShare")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func loadShare(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitNoCheck(r)

	uid := req.GetParamInt("uid")
	title := req.GetParamString("title")
	thumbnail := req.GetParamStringDef("thumbnail", "")
	img := req.GetParamStringDef("img", "")
	dst := req.GetParamString("dst")
	origin := req.GetParamInt("origin")
	tags := req.GetParamIntArray("tags")
	width := req.GetParamIntDef("width", 0)
	height := req.GetParamIntDef("height", 0)
	src := req.GetParamStringDef("src", "")
	cdn := req.GetParamStringDef("cdn", "")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, uid, "AddShare",
		&share.ShareRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Title: title, Img: img, Thumbnail: thumbnail,
			Dst: dst, Tags: tags, Origin: origin, Src: src, Cdn: cdn,
			Width: width, Height: height})

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

	id := req.GetParamInt("id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "Reshare",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id})

	httpserver.CheckRPCErr(rpcerr, "Reshare")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "Reshare")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func unshare(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	id := req.GetParamInt("id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "Unshare",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id})

	httpserver.CheckRPCErr(rpcerr, "Unshare")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "Unshare")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addComment(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	id := req.GetParamInt("id")
	content := req.GetParamString("content")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "AddComment",
		&share.CommentRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id, Content: content})

	httpserver.CheckRPCErr(rpcerr, "AddComment")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddComment")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getUserShares(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")
	tuid := req.GetParamInt("tuid")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "GetUserShares",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Seq: seq, Num: num, Id: tuid})

	httpserver.CheckRPCErr(rpcerr, "GetUserShares")
	res := resp.Interface().(*share.ShareReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetUserShares")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getShares(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")
	id := req.GetParamInt("tag_id")
	log.Printf("getShares request:%v", r)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "GetShares",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Seq: seq, Num: num, Id: id})

	httpserver.CheckRPCErr(rpcerr, "GetShares")
	res := resp.Interface().(*share.ShareReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetShares")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getShareComments(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	id := req.GetParamInt("id")
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "GetShareComments",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id, Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "GetShareComments")
	res := resp.Interface().(*share.CommentReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetShareComments")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getShareIds(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")
	tag := req.GetParamInt("tag_id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "GetShareIds",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: tag, Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "GetShareIds")
	res := resp.Interface().(*share.ShareIdReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetShareIds")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getRecommendShares(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	tag := req.GetParamInt("tag_id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "GetRecommendShares",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: tag})

	httpserver.CheckRPCErr(rpcerr, "GetRecommendShares")
	res := resp.Interface().(*share.RecommendShareReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetRecommendShares")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getUserInfo(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	tuid := req.GetParamInt("tuid")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.UserServerType, req.Uid, "GetInfo",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid}, Id: tuid})

	httpserver.CheckRPCErr(rpcerr, "GetInfo")
	res := resp.Interface().(*user.InfoReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetInfo")

	body := httpserver.GenInfoResponseBody(res)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func modUserInfo(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	headurl := req.GetParamStringDef("headurl", "")
	nickname := req.GetParamStringDef("nickname", "")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.UserServerType, req.Uid, "ModInfo",
		&user.ModInfoRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Headurl: headurl, Nickname: nickname})

	httpserver.CheckRPCErr(rpcerr, "ModInfo")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "ModInfo")

	body := httpserver.GenInfoResponseBody(res)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getShareDetail(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	id := req.GetParamInt("id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "GetShareDetail",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid}, Id: id})

	httpserver.CheckRPCErr(rpcerr, "GetShareDetail")
	res := resp.Interface().(*share.ShareDetailReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "GetShareDetail")

	body := httpserver.GenInfoResponseBody(res)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func reportClick(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	dev := req.ParseDevice(r)

	id := req.GetParamInt("id")
	ctype := req.GetParamInt("type")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ModifyServerType, req.Uid, "ReportClick",
		&modify.ClickRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Type: ctype, Id: id, Imei: dev.Imei})

	httpserver.CheckRPCErr(rpcerr, "ReportClick")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "ReportClick")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func report(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)

	id := req.GetParamInt("id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ModifyServerType, req.Uid, "Report",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id})

	httpserver.CheckRPCErr(rpcerr, "Report")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "Report")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func checkUpdate(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
	dev := req.ParseDevice(r)
	var term int64
	if dev.Model == "iPhone" {
		term = 1
	}

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "CheckUpdate",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid,
			Term: term, Version: dev.Version}})

	httpserver.CheckRPCErr(rpcerr, "CheckUpdate")
	res := resp.Interface().(*config.VersionReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "CheckUpdate")

	body := httpserver.GenResponseBody(res, false)
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
	mux.Handle("/get_user_shares", httpserver.AppHandler(getUserShares))
	mux.Handle("/get_shares", httpserver.AppHandler(getShares))
	mux.Handle("/get_share_comments", httpserver.AppHandler(getShareComments))
	mux.Handle("/get_share_ids", httpserver.AppHandler(getShareIds))
	mux.Handle("/get_share_detail", httpserver.AppHandler(getShareDetail))
	mux.Handle("/get_recommend_shares", httpserver.AppHandler(getRecommendShares))
	mux.Handle("/reshare", httpserver.AppHandler(reshare))
	mux.Handle("/unshare", httpserver.AppHandler(unshare))
	mux.Handle("/add_comment", httpserver.AppHandler(addComment))
	mux.Handle("/get_user_info", httpserver.AppHandler(getUserInfo))
	mux.Handle("/mod_user_info", httpserver.AppHandler(modUserInfo))
	mux.Handle("/report_click", httpserver.AppHandler(reportClick))
	mux.Handle("/report", httpserver.AppHandler(report))
	mux.Handle("/check_update", httpserver.AppHandler(checkUpdate))
	mux.Handle("/load_share", httpserver.AppHandler(loadShare))
	return mux
}
