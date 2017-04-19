package main

import (
	"Server/util"
	"log"
	"net/http"

	"github.com/facebookgo/grace/gracehttp"
)

func init() {
	w := util.NewRotateWriter("/data/server/laughing.log", 1024*1024*1024)
	log.SetOutput(w)
}

func main() {
	gracehttp.Serve(
		&http.Server{Addr: ":81", Handler: NewAppServer()},
	)
}
