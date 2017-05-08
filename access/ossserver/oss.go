package main

import (
	"laughing-server/httpserver"
	"laughing-server/proto/common"
	"laughing-server/proto/config"
	"laughing-server/proto/share"
	"laughing-server/proto/user"
	"laughing-server/proto/verify"
	"laughing-server/util"
	"net/http"
)

func login(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitNoCheck(r)
	username := req.GetParamString("username")
	passwd := req.GetParamString("passwd")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.VerifyServerType, 0, "BackLogin",
		&verify.BackLoginRequest{Head: &common.Head{Sid: uuid},
			Username: username, Passwd: passwd})

	httpserver.CheckRPCErr(rpcerr, "BackLogin")
	res := resp.Interface().(*verify.LoginReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "BackLogin")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getTags(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "FetchTags",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "FetchTags")
	res := resp.Interface().(*share.TagReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FetchTags")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addTag(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	content := req.GetParamString("content")
	img := req.GetParamString("img")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "AddTag",
		&share.TagRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &share.TagInfo{Content: content, Img: img}})

	httpserver.CheckRPCErr(rpcerr, "AddTags")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddTag")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func delTags(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	ids := req.GetParamWrapIntArray("ids")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "DelTags",
		&share.DelTagRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Ids: ids})

	httpserver.CheckRPCErr(rpcerr, "DelTags")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "DelTag")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getVersions(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "FetchVersions",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "FetchVersions")
	res := resp.Interface().(*config.FetchVersionReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FetchVersions")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addVersion(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	term := req.GetParamInt("term")
	version := req.GetParamInt("version")
	vname := req.GetParamString("vname")
	title := req.GetParamString("title")
	subtitle := req.GetParamString("subtitle")
	downurl := req.GetParamString("downurl")
	desc := req.GetParamString("desc")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "AddVersion",
		&config.VersionRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &config.VersionInfo{Term: term, Version: version,
				Vname: vname, Title: title, Subtitle: subtitle, Downurl: downurl,
				Desc: desc}})

	httpserver.CheckRPCErr(rpcerr, "AddVersion")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddVersion")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getUsers(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.UserServerType, req.Uid, "FetchInfos",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "FetchInfos")
	res := resp.Interface().(*user.FetchInfosReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FetchInfos")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addUser(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	nickname := req.GetParamString("nickname")
	headurl := req.GetParamString("headurl")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.UserServerType, req.Uid, "AddInfo",
		&user.InfoRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &user.Info{Nickname: nickname, Headurl: headurl}})

	httpserver.CheckRPCErr(rpcerr, "AddInfo")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddInfo")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getShares(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")
	rtype := req.GetParamInt("type")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "FetchShares",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Seq: seq, Num: num, Type: rtype})

	httpserver.CheckRPCErr(rpcerr, "FetchShares")
	res := resp.Interface().(*share.ShareReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FetchShares")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addShareTags(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	id := req.GetParamInt("id")
	tags := req.GetParamWrapIntArray("tags")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "AddShareTags",
		&share.AddTagRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id, Tags: tags})

	httpserver.CheckRPCErr(rpcerr, "AddShareTags")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddShareTags")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func reviewShare(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	id := req.GetParamInt("id")
	reject := req.GetParamIntDef("reject", 0)
	modify := req.GetParamIntDef("modify", 0)
	title := req.GetParamStringDef("title", "")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "ReviewShare",
		&share.ReviewShareRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id, Reject: reject, Modify: modify, Title: title})

	httpserver.CheckRPCErr(rpcerr, "ReviewShare")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "ReviewShare")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

//NewOssServer return oss http handler
func NewOssServer() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/login", httpserver.AppHandler(login))
	mux.Handle("/get_tags", httpserver.AppHandler(getTags))
	mux.Handle("/add_tag", httpserver.AppHandler(addTag))
	mux.Handle("/del_tags", httpserver.AppHandler(delTags))
	mux.Handle("/add_version", httpserver.AppHandler(addVersion))
	mux.Handle("/get_versions", httpserver.AppHandler(getVersions))
	mux.Handle("/get_shares", httpserver.AppHandler(getShares))
	mux.Handle("/get_users", httpserver.AppHandler(getUsers))
	mux.Handle("/add_user", httpserver.AppHandler(addUser))
	mux.Handle("/review_share", httpserver.AppHandler(reviewShare))
	mux.Handle("/add_share_tags", httpserver.AppHandler(addShareTags))
	return mux
}
