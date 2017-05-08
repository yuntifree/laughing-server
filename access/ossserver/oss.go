package main

import (
	"Server/httpserver"
	"Server/proto/common"
	"Server/util"
	"net/http"
)

func login(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.Init(r)
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

//NewAppServer return app http handler
func NewAppServer() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/login", httpserver.AppHandler(login))
	return mux
}
