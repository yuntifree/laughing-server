package main

import (
	"laughing-server/httpserver"
	"laughing-server/proto/common"
	"laughing-server/proto/share"
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

//NewOssServer return oss http handler
func NewOssServer() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/login", httpserver.AppHandler(login))
	mux.Handle("/get_tags", httpserver.AppHandler(getTags))
	mux.Handle("/add_tag", httpserver.AppHandler(addTag))
	mux.Handle("/del_tags", httpserver.AppHandler(delTags))
	return mux
}
