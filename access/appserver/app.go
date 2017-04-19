package main

import (
	"laughing-server/httpserver"
	"laughing-server/proto/common"
	"laughing-server/proto/fan"
	"laughing-server/util"
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

//NewAppServer return app http handler
func NewAppServer() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/followop", httpserver.AppHandler(followop))
	return mux
}
