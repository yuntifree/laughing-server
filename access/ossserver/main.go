package main

import (
	"laughing-server/util"
	"log"
	"net/http"
	"time"

	"github.com/facebookgo/grace/gracehttp"
)

func init() {
	w := util.NewRotateWriter("/data/server/laughoss.log", 1024*1024*1024)
	log.SetOutput(w)
}

func main() {
	gracehttp.Serve(
		&http.Server{Addr: ":8089", Handler: NewOssServer(), IdleTimeout: 30 * time.Second},
	)
}
