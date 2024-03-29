package httpserver

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"laughing-server/proto/common"
	"laughing-server/proto/discover"
	"laughing-server/proto/limit"
	"laughing-server/proto/verify"
	"laughing-server/util"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	simplejson "github.com/bitly/go-simplejson"
	nsq "github.com/nsqio/go-nsq"
)

const (
	//ErrOk success code
	ErrOk = iota
	//ErrMissParam miss parameter
	ErrMissParam
	//ErrInvalidParam invalid parameters
	ErrInvalidParam
	//ErrDatabase database operation failed
	ErrDatabase
	//ErrInner some unexpected inner failure
	ErrInner
	//ErrPanic some unexpected panic
	ErrPanic
)
const (
	//ErrToken illegal token
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
		log.Printf("report request api:%s failed:%v", method, err)
	}
	return
}

//ReportSuccResp report success response
func ReportSuccResp(uri string) {
	method := extractAPIName(uri)
	err := util.PubResponse(w, method, 0)
	if err != nil {
		log.Printf("report response api:%s failed:%v", method, err)
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

func getJSONStringArray(js *simplejson.Json, key string) []string {
	var s []string
	arr, err := js.Get(key).Array()
	if err == nil {
		for i := 0; i < len(arr); i++ {
			v, err := js.Get(key).GetIndex(i).String()
			if err != nil {
				log.Printf("getJSONIntArray get failed, idx:%d %v", i, err)
				continue
			}
			s = append(s, v)
		}
	}
	return s
}

func getWrapJSONStringArray(js *simplejson.Json, key string) []string {
	var s []string
	arr, err := js.Get("data").Get(key).Array()
	if err == nil {
		for i := 0; i < len(arr); i++ {
			v, err := js.Get("data").Get(key).GetIndex(i).String()
			if err != nil {
				log.Printf("getJSONIntArray get failed, idx:%d %v", i, err)
				continue
			}
			s = append(s, v)
		}
	}
	return s
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

func getJSONIntArray(js *simplejson.Json, key string) []int64 {
	var s []int64
	arr, err := js.Get(key).Array()
	if err == nil {
		for i := 0; i < len(arr); i++ {
			v, err := js.Get(key).GetIndex(i).Int64()
			if err != nil {
				log.Printf("getJSONIntArray get failed, idx:%d %v", i, err)
				continue
			}
			s = append(s, v)
		}
	}
	return s
}

func getWrapJSONIntArray(js *simplejson.Json, key string) []int64 {
	var s []int64
	arr, err := js.Get("data").Get(key).Array()
	if err == nil {
		for i := 0; i < len(arr); i++ {
			v, err := js.Get("data").Get(key).GetIndex(i).Int64()
			if err != nil {
				log.Printf("getJSONIntArray get failed, idx:%d %v", i, err)
				continue
			}
			s = append(s, v)
		}
	}
	return s
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
	Uid      int64
	Token    string
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

func checkNonce(uid int64, nonce string) bool {
	uuid := util.GenUUID()
	resp, err := CallRPC(util.LimitServerType, uid, "CheckDuplicate",
		&limit.CheckDuplicateRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Type: limit.CheckType_NONCE, Id: nonce})
	if err.Interface() != nil {
		return false
	}
	res := resp.Interface().(*common.CommReply)
	if res.Head.Retcode != 0 {
		log.Printf("check duplicate nonce failed:%d %s", uid, nonce)
		return false
	}

	return true
}

func checkToken(uid int64, token string) bool {
	uuid := util.GenUUID()
	resp, err := CallRPC(util.VerifyServerType, uid, "CheckToken",
		&verify.CheckTokenRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Token: token})
	if err.Interface() != nil {
		return false
	}
	res := resp.Interface().(*common.CommReply)
	if res.Head.Retcode != 0 {
		log.Printf("check token failed:%d %s", uid, token)
		return false
	}

	return true
}

func checkBackToken(uid int64, token string) bool {
	uuid := util.GenUUID()
	resp, err := CallRPC(util.VerifyServerType, uid, "CheckBackToken",
		&verify.CheckTokenRequest{Head: &common.Head{Sid: uuid, Uid: uid},
			Token: token})
	if err.Interface() != nil {
		return false
	}
	res := resp.Interface().(*common.CommReply)
	if res.Head.Retcode != 0 {
		log.Printf("check token failed:%d %s", uid, token)
		return false
	}

	return true
}

//InitNoCheck init without any check
func (r *Request) InitNoCheck(req *http.Request) {
	ReportRequest(req.RequestURI)
	var err error
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Printf("ReadAll failed:%v", err)
		panic(util.AppError{ErrInvalidParam, "invalid param", r.Callback})
	}
	log.Printf("body buf:%s", string(buf))
	r.Post, err = simplejson.NewJson(buf)
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

//CheckOssCookie check oss cookie
func CheckOssCookie(req *http.Request) bool {
	var uid int64
	var token string
	c1, err := req.Cookie("u")
	if err == nil {
		id, _ := strconv.Atoi(c1.Value)
		uid = int64(id)
	}
	c2, err := req.Cookie("s")
	if err == nil {
		token = c2.Value
	}
	if !checkBackToken(uid, token) {
		return false
	}
	return true
}

//InitOss init oss request
func (r *Request) InitOss(req *http.Request) {
	r.InitNoCheck(req)
	r.Uid = getJSONInt(r.Post, "uid")
	if r.Uid == 0 {
		panic(util.AppError{ErrInvalidParam, "need login", r.Callback})
	}
	r.Token = getJSONString(r.Post, "token")
	if !checkBackToken(r.Uid, r.Token) {
		panic(util.AppError{ErrInvalidParam, "illegal token", r.Callback})
	}
}

//Init init request
func (r *Request) Init(req *http.Request) {
	r.InitNoCheck(req)
	c1, err := req.Cookie("u")
	if err == nil {
		id, _ := strconv.Atoi(c1.Value)
		r.Uid = int64(id)
	}
	c2, err := req.Cookie("s")
	if err == nil {
		r.Token = c2.Value
	}
	nonce := getJSONString(r.Post, "nonce")
	if !checkNonce(r.Uid, nonce) {
		log.Printf("checkNonce failed:%d %s", r.Uid, nonce)
		panic(util.AppError{ErrInvalidParam, "duplicate nonce", r.Callback})
	}
}

//InitCheck init request and check token
func (r *Request) InitCheck(req *http.Request) {
	r.Init(req)
	if r.Uid == 0 {
		panic(util.AppError{ErrInvalidParam, "need login", r.Callback})
	}
	if !checkToken(r.Uid, r.Token) {
		panic(util.AppError{ErrInvalidParam, "illegal token", r.Callback})
	}
}

//DeviceInfo device info
type DeviceInfo struct {
	Imei     string
	Model    string
	Language string
	Version  int64
	Os       string
	Api      string
	Wifi     int64
}

func getArrStr(arr []string) string {
	if len(arr) > 0 {
		return arr[0]
	}
	return ""
}

//ParseDevice parse device info
func (r *Request) ParseDevice(req *http.Request) (info DeviceInfo) {
	device := req.Header.Get("X-Cc-Device")
	if device == "" {
		return
	}
	m, _ := url.ParseQuery(device)
	info.Imei = getArrStr(m["imei"])
	info.Model = getArrStr(m["model"])
	info.Language = getArrStr(m["language"])
	info.Version, _ = strconv.ParseInt(getArrStr(m["version"]), 10, 64)
	info.Os = getArrStr(m["os"])
	info.Api = getArrStr(m["api"])
	info.Wifi, _ = strconv.ParseInt(getArrStr(m["wifi"]), 10, 64)
	return
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

func (r *Request) GetParamIntArray(key string) []int64 {
	return getJSONIntArray(r.Post, key)
}
func (r *Request) GetParamWrapIntArray(key string) []int64 {
	return getWrapJSONIntArray(r.Post, key)
}

func (r *Request) GetParamStringArray(key string) []string {
	return getJSONStringArray(r.Post, key)
}

func (r *Request) GetParamWrapStringArray(key string) []string {
	return getWrapJSONStringArray(r.Post, key)
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
		js.Set("desc", "please retry later")
	} else if e.Code < ErrToken {
		js.Set("desc", "please retry later")
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

//GenInfoResponseBody generate info response body
func GenInfoResponseBody(res interface{}) []byte {
	js, err := simplejson.NewJson([]byte(`{"errno":0}`))
	if err != nil {
		panic(util.AppError{ErrInner, err.Error(), ""})
	}
	val := reflect.ValueOf(res).Elem()
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		if typeField.Name == "Info" {
			js.Set("data", valueField.Interface())
			break
		}
	}
	data, err := js.MarshalJSON()
	if err != nil {
		panic(util.AppError{ErrInner, err.Error(), ""})
	}

	return data
}

//GenResponseBody generate response body
func GenResponseBody(res interface{}, flag bool) []byte {
	return GenResponseBodyCallback(res, "", flag)
}

var nilStrings = []string{
	"Infos",
	"Recommendtag",
}

func isNilStrings(name string) bool {
	for _, v := range nilStrings {
		if v == name {
			return true
		}
	}
	return false
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
		} else if isNilStrings(typeField.Name) {
			if valueField.IsNil() {
				continue
			} else {
				js.SetPath([]string{"data", strings.ToLower(typeField.Name)},
					valueField.Interface())
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

//CallRPC call rpc method
func CallRPC(rtype, uid int64, method string, request interface{}) (reflect.Value, reflect.Value) {
	return CallRPCCallback(rtype, uid, method, "", request)
}

//CallRPC call rpc method with callback
func CallRPCCallback(rtype, uid int64, method, callback string, request interface{}) (reflect.Value, reflect.Value) {
	var resp reflect.Value
	serverName := util.GenServerName(rtype, callback)
	address := GetNameServerCallback(uid, serverName, callback)
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return resp, reflect.ValueOf(err)
	}
	defer conn.Close()
	cli := util.GenClient(rtype, conn, callback)
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
