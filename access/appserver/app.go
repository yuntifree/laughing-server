package main

import (
	"laughing-server/httpserver"
	"laughing-server/proto/common"
	"laughing-server/proto/fan"
	"laughing-server/proto/verify"
	"laughing-server/util"
	"log"
	"net/http"
)

const (
	FollowType   = 0
	UnfollowType = 1
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

//NewAppServer return app http handler
func NewAppServer() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/followop", httpserver.AppHandler(followop))
	mux.Handle("/fblogin", httpserver.AppHandler(fblogin))
	return mux
}
