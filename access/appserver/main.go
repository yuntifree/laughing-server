package main

import (
	"laughing-server/util"
	"log"
	"net/http"
	"time"

	"github.com/facebookgo/grace/gracehttp"
)

func init() {
	w := util.NewRotateWriter("/data/server/app.log", 1024*1024*1024)
	log.SetOutput(w)
}

func main() {
	gracehttp.Serve(
		&http.Server{Addr: ":8088", Handler: NewAppServer(), IdleTimeout: 30 * time.Second},
	)
}
