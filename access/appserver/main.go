package main

import (
	"flag"
	"laughing-server/util"
	"log"
	"net/http"
	"time"

	"github.com/facebookgo/grace/gracehttp"
	"github.com/gorilla/mux"
)

func init() {
	w := util.NewRotateWriter("/data/server/app.log", 1024*1024*1024)
	log.SetOutput(w)
}

func handleFile(r *mux.Router, root string) *mux.Router {
	r.Handle("/", http.FileServer(http.Dir(root)))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/",
		http.FileServer(http.Dir(root+"/css/"))))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/",
		http.FileServer(http.Dir(root+"/images/"))))
	return r
}

func main() {
	addr := flag.String("addr", ":8088", "bind address")
	root := flag.String("root", "/data/server", "root directory")
	flag.Parse()
	r := NewAppServer()
	r = handleFile(r, *root)
	gracehttp.Serve(
		&http.Server{Addr: *addr, Handler: r, IdleTimeout: 30 * time.Second},
	)
}
