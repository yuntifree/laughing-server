package httpserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"laughing-server/proto/common"
	"laughing-server/proto/discover"
	"laughing-server/proto/fan"
	"laughing-server/proto/verify"
	"laughing-server/util"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	simplejson "github.com/bitly/go-simplejson"
	nsq "github.com/nsqio/go-nsq"
)

const (
	ErrOk = iota
	ErrMissParam
	ErrInvalidParam
	ErrDatabase
	ErrInner
	ErrPanic
)
const (
	ErrToken = iota + 101
)

var w *nsq.Producer

func init() {
	w = util.NewNsqProducer()
}

func extractAPIName(uri string) string {
	pos := strings.Index(uri, "?")
	path := uri
	if pos != -1 {
		path = uri[0:pos]
	}
	lpos := strings.LastIndex(path, "/")
	method := path
	if lpos != -1 {
		method = path[lpos+1:]
	}
	return method
}

//ReportRequest report request
func ReportRequest(uri string) {
	method := extractAPIName(uri)
	err := util.PubRequest(w, method)
	if err != nil {
		log.Printf("report request api:%s failed:%v", err)
	}
	return
}

//ReportSuccResp report success response
func ReportSuccResp(uri string) {
	method := extractAPIName(uri)
	err := util.PubResponse(w, method, 0)
	if err != nil {
		log.Printf("report response api:%s failed:%v", err)
	}
	return
}

func genParamErr(key string) string {
	return "get param:" + key + " failed"
}

func getJSONString(js *simplejson.Json, key string) string {
	if val, err := js.Get(key).String(); err == nil {
		return val
	}

	if val, err := js.Get("data").Get(key).String(); err == nil {
		return val
	}
	panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key)})
}

func getJSONStringDef(js *simplejson.Json, key, def string) string {
	if val, err := js.Get(key).String(); err == nil {
		return val
	}

	if val, err := js.Get("data").Get(key).String(); err == nil {
		return val
	}
	return def
}

func getJSONInt(js *simplejson.Json, key string) int64 {
	if val, err := js.Get(key).Int64(); err == nil {
		return val
	}

	if val, err := js.Get("data").Get(key).Int64(); err == nil {
		return val
	}
	panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key)})
}

func getJSONIntDef(js *simplejson.Json, key string, def int64) int64 {
	if val, err := js.Get(key).Int64(); err == nil {
		return val
	}

	if val, err := js.Get("data").Get(key).Int64(); err == nil {
		return val
	}
	return def
}

func getJSONBool(js *simplejson.Json, key string) bool {
	if val, err := js.Get(key).Bool(); err == nil {
		return val
	}

	if val, err := js.Get("data").Get(key).Bool(); err == nil {
		return val
	}
	panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key)})
}

func getJSONBoolDef(js *simplejson.Json, key string, def bool) bool {
	if val, err := js.Get(key).Bool(); err == nil {
		return val
	}

	if val, err := js.Get("data").Get(key).Bool(); err == nil {
		return val
	}
	return def
}

func getJSONFloat(js *simplejson.Json, key string) float64 {
	if val, err := js.Get(key).Float64(); err == nil {
		return val
	}

	if val, err := js.Get("data").Get(key).Float64(); err == nil {
		return val
	}
	panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key)})
}

func getJSONFloatDef(js *simplejson.Json, key string, def float64) float64 {
	if val, err := js.Get(key).Float64(); err == nil {
		return val
	}

	if val, err := js.Get("data").Get(key).Float64(); err == nil {
		return val
	}
	return def
}

func getFormInt(v url.Values, key, callback string) int64 {
	vals := v[key]
	if len(vals) == 0 {
		panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key),
			Callback: callback})
	}
	val, err := strconv.ParseInt(vals[0], 10, 64)
	if err != nil {
		panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key),
			Callback: callback})
	}
	return val
}

func getFormIntDef(v url.Values, key string, def int64) int64 {
	vals := v[key]
	if len(vals) == 0 {
		return def
	}
	val, err := strconv.ParseInt(vals[0], 10, 64)
	if err != nil {
		return def
	}
	return val
}

func getFormFloat(v url.Values, key, callback string) float64 {
	vals := v[key]
	if len(vals) == 0 {
		panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key),
			Callback: callback})
	}
	val, err := strconv.ParseFloat(vals[0], 64)
	if err != nil {
		panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key),
			Callback: callback})
	}
	return val
}

func getFormFloatDef(v url.Values, key string, def float64) float64 {
	vals := v[key]
	if len(vals) == 0 {
		return def
	}
	val, err := strconv.ParseFloat(vals[0], 64)
	if err != nil {
		return def
	}
	return val
}

func getFormBool(v url.Values, key, callback string) bool {
	vals := v[key]
	if len(vals) == 0 {
		panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key),
			Callback: callback})
	}
	val, err := strconv.ParseBool(vals[0])
	if err != nil {
		panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key),
			Callback: callback})
	}
	return val
}

func getFormBoolDef(v url.Values, key string, def bool) bool {
	vals := v[key]
	if len(vals) == 0 {
		return def
	}
	val, err := strconv.ParseBool(vals[0])
	if err != nil {
		return def
	}
	return val
}

func getFormString(v url.Values, key, callback string) string {
	vals := v[key]
	if len(vals) == 0 {
		panic(util.AppError{Code: ErrMissParam, Msg: genParamErr(key),
			Callback: callback})
	}
	return vals[0]
}

func getFormStringDef(v url.Values, key string, def string) string {
	vals := v[key]
	if len(vals) == 0 {
		return def
	}
	return vals[0]
}

//Request request infos
type Request struct {
	Post     *simplejson.Json
	Form     url.Values
	debug    bool
	Callback string
}

func writeRsp(w http.ResponseWriter, body []byte, callback string) {
	if callback != "" {
		var buf bytes.Buffer
		buf.Write([]byte(callback))
		buf.Write([]byte("("))
		buf.Write(body)
		buf.Write([]byte(")"))
		w.Write(buf.Bytes())
		return
	}
	w.Write(body)
	return
}

//WriteRsp support for callback
func (r *Request) WriteRsp(w http.ResponseWriter, body []byte) {
	writeRsp(w, body, r.Callback)
}

//Init init request
func (r *Request) Init(req *http.Request) {
	ReportRequest(req.RequestURI)
	var err error
	r.Post, err = simplejson.NewFromReader(req.Body)
	if err == io.EOF {
		req.ParseForm()
		r.Form = req.Form
		r.debug = true
		r.Callback = getFormString(r.Form, "callback", "")
		return
	}
	if err != nil {
		log.Printf("parse reqbody failed:%v", err)
		panic(util.AppError{ErrInvalidParam, "invalid param", r.Callback})
	}
}

func (r *Request) GetParamInt(key string) int64 {
	if r.debug {
		return getFormInt(r.Form, key, r.Callback)
	}
	return getJSONInt(r.Post, key)
}

func (r *Request) GetParamIntDef(key string, def int64) int64 {
	if r.debug {
		return getFormIntDef(r.Form, key, def)
	}
	return getJSONIntDef(r.Post, key, def)
}

func (r *Request) GetParamBool(key string) bool {
	if r.debug {
		return getFormBool(r.Form, key, r.Callback)
	}
	return getJSONBool(r.Post, key)
}

func (r *Request) GetParamBoolDef(key string, def bool) bool {
	if r.debug {
		return getFormBoolDef(r.Form, key, def)
	}
	return getJSONBoolDef(r.Post, key, def)
}

func (r *Request) GetParamString(key string) string {
	if r.debug {
		return getFormString(r.Form, key, r.Callback)
	}
	return getJSONString(r.Post, key)
}
func (r *Request) GetParamStringDef(key string, def string) string {
	if r.debug {
		return getFormStringDef(r.Form, key, def)
	}
	return getJSONStringDef(r.Post, key, def)
}

func (r *Request) GetParamFloat(key string) float64 {
	if r.debug {
		return getFormFloat(r.Form, key, r.Callback)
	}
	return getJSONFloat(r.Post, key)
}
func (r *Request) GetParamFloatDef(key string, def float64) float64 {
	if r.debug {
		return getFormFloatDef(r.Form, key, def)
	}
	return getJSONFloatDef(r.Post, key, def)
}

func extractError(r interface{}) *util.AppError {
	if k, ok := r.(util.AppError); ok {
		return &k
	}
	log.Printf("unexpected panic:%v", r)
	return &util.AppError{ErrPanic, r.(error).Error(), ""}
}

func handleError(w http.ResponseWriter, e *util.AppError) {
	log.Printf("error code:%d msg:%s callback:%s", e.Code, e.Msg,
		e.Callback)

	js, _ := simplejson.NewJson([]byte(`{}`))
	js.Set("errno", e.Code)
	if e.Code == ErrInvalidParam || e.Code == ErrMissParam {
		js.Set("errno", ErrToken)
		js.Set("desc", "服务器又傲娇了~")
	} else if e.Code < ErrToken {
		js.Set("desc", "服务器又傲娇了~")
	} else {
		js.Set("desc", e.Msg)
	}
	body, err := js.MarshalJSON()
	if err != nil {
		log.Printf("MarshalJSON failed: %v", err)
		writeRsp(w, []byte(`{"errno":2,"desc":"invalid param"}`), e.Callback)
		return
	}
	writeRsp(w, body, e.Callback)
}

type AppHandler func(http.ResponseWriter, *http.Request) *util.AppError

func (fn AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			apperr := extractError(r)
			handleError(w, apperr)
		}
	}()
	if e := fn(w, r); e != nil {
		handleError(w, e)
	}
}

func getDiscoverAddress() string {
	ip := util.GetInnerIP()
	if ip != util.DebugHost {
		hosts := strings.Split(util.APIHosts, ",")
		if len(hosts) > 0 {
			idx := util.Randn(int32(len(hosts)))
			return hosts[idx] + util.DiscoverServerPort
		}
	}
	return "localhost" + util.DiscoverServerPort
}

//GetNameServer
func GetNameServer(uid int64, name string) string {
	return GetNameServerCallback(uid, name, "")
}

//GetNameServerCallback get server from name service with callback
func GetNameServerCallback(uid int64, name, callback string) string {
	address := getDiscoverAddress()
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Printf("did not connect %s: %v", address, err)
		panic(util.AppError{ErrInner, err.Error(), callback})
	}
	defer conn.Close()
	c := discover.NewDiscoverClient(conn)

	ip := util.GetInnerIP()
	if ip == util.DebugHost {
		name += ":debug"
	}
	uuid := util.GenUUID()
	res, err := c.Resolve(context.Background(),
		&discover.ServerRequest{Head: &common.Head{Uid: uid, Sid: uuid}, Sname: name})
	if err != nil {
		log.Printf("Resolve failed %s: %v", name, err)
		panic(util.AppError{ErrInner, err.Error(), callback})
	}

	if res.Head.Retcode != 0 {
		log.Printf("Resolve failed  name:%s errcode:%d\n", name, res.Head.Retcode)
		panic(util.AppError{ErrInner,
			fmt.Sprintf("Resolve failed  name:%s errcode:%d\n", name, res.Head.Retcode), callback})
	}

	return res.Host
}

//RspGzip response with gzip
func RspGzip(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", "application/json")
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write(body)
	gw.Close()
	w.Write(buf.Bytes())
}

//GenResponseBody generate response body
func GenResponseBody(res interface{}, flag bool) []byte {
	return GenResponseBodyCallback(res, "", flag)
}

//GenResponseBodyCallback generate response body with callback
func GenResponseBodyCallback(res interface{}, callback string, flag bool) []byte {
	js, err := simplejson.NewJson([]byte(`{"errno":0}`))
	if err != nil {
		panic(util.AppError{ErrInner, err.Error(), callback})
	}
	val := reflect.ValueOf(res).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		if typeField.Name == "Head" {
			if flag {
				headVal := reflect.Indirect(valueField)
				uid := headVal.FieldByName("Uid")
				js.SetPath([]string{"data", "uid"}, uid.Interface())

			} else {
				continue
			}
		} else {
			js.SetPath([]string{"data", strings.ToLower(typeField.Name)},
				valueField.Interface())
		}
	}
	data, err := js.MarshalJSON()
	if err != nil {
		panic(util.AppError{ErrInner, err.Error(), callback})
	}

	return data
}

//CheckRPCErr check rpc response error
func CheckRPCErr(err reflect.Value, method string) {
	CheckRPCErrCallback(err, method, "")
	return
}

func CheckRPCErrCallback(err reflect.Value, method, callback string) {
	if err.Interface() != nil {
		log.Printf("RPC %s failed:%v", method, err)
		panic(util.AppError{ErrInner, "grpc failed " + method, callback})
	}
}

//CheckRPCCode check rpc response code
func CheckRPCCode(retcode common.ErrCode, method string) {
	CheckRPCCodeCallback(retcode, method, "")
	return
}

//CheckRPCCodeCallback check rpc response code with callback
func CheckRPCCodeCallback(retcode common.ErrCode, method, callback string) {
	if retcode != 0 {
		log.Printf("%s failed retcode:%d", method, retcode)
	}
	if retcode == common.ErrCode_INVALID_TOKEN {
		panic(util.AppError{ErrToken, "token验证失败", callback})
	} else if retcode != 0 {
		panic(util.AppError{int(retcode), "服务器又傲娇了~", callback})
	}
}

func genServerName(rtype int64, callback string) string {
	switch rtype {
	case util.FanServerType:
		return util.FanServerName
	case util.VerifyServerType:
		return util.VerifyServerName
	default:
		panic(util.AppError{ErrInvalidParam, "illegal server type", callback})
	}
}

func genClient(rtype int64, conn *grpc.ClientConn, callback string) interface{} {
	var cli interface{}
	switch rtype {
	case util.FanServerType:
		cli = fan.NewFanClient(conn)
	case util.VerifyServerType:
		cli = verify.NewVerifyClient(conn)
	default:
		panic(util.AppError{ErrInvalidParam, "illegal server type", callback})
	}
	return cli
}

//CallRPC call rpc method
func CallRPC(rtype, uid int64, method string, request interface{}) (reflect.Value, reflect.Value) {
	return CallRPCCallback(rtype, uid, method, "", request)
}

//CallRPC call rpc method with callback
func CallRPCCallback(rtype, uid int64, method, callback string, request interface{}) (reflect.Value, reflect.Value) {
	var resp reflect.Value
	serverName := genServerName(rtype, callback)
	address := GetNameServerCallback(uid, serverName, callback)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return resp, reflect.ValueOf(err)
	}
	defer conn.Close()
	cli := genClient(rtype, conn, callback)
	ctx := context.Background()

	inputs := make([]reflect.Value, 2)
	inputs[0] = reflect.ValueOf(ctx)
	inputs[1] = reflect.ValueOf(request)
	arr := reflect.ValueOf(cli).MethodByName(method).Call(inputs)
	if len(arr) != 2 {
		log.Printf("callRPC arr len%d", len(arr))
		return resp, reflect.ValueOf(errors.New("illegal grpc call response"))
	}
	return arr[0], arr[1]
}

//FileHandler wrapper for FileServer
type FileHandler struct {
	Dir string
	h   http.Handler
}

//NewFileHandler return new FileHandler
func NewFileHandler(dir string) *FileHandler {
	return &FileHandler{
		Dir: dir,
		h:   http.FileServer(http.Dir(dir)),
	}
}

//ServeHTTP FileHandler implemention
func (f FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("url:%s", r.URL)
	f.h.ServeHTTP(w, r)
}
