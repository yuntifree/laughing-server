package main

import (
	"io/ioutil"
	"laughing-server/httpserver"
	"laughing-server/proto/common"
	"laughing-server/proto/config"
	"laughing-server/proto/share"
	"laughing-server/proto/user"
	"laughing-server/proto/verify"
	"laughing-server/ucloud"
	"laughing-server/util"
	"log"
	"net/http"

	simplejson "github.com/bitly/go-simplejson"
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
	recommend := req.GetParamIntDef("recommend", 0)
	hot := req.GetParamIntDef("hot", 0)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "AddTag",
		&share.TagRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &share.TagInfo{Content: content, Img: img,
				Recommend: recommend, Hot: hot}})

	httpserver.CheckRPCErr(rpcerr, "AddTags")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddTag")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func modTag(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	content := req.GetParamString("content")
	img := req.GetParamString("img")
	recommend := req.GetParamInt("recommend")
	hot := req.GetParamInt("hot")
	priority := req.GetParamInt("priority")
	id := req.GetParamInt("id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "ModTag",
		&share.TagRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &share.TagInfo{Content: content, Img: img,
				Recommend: recommend, Id: id, Hot: hot, Priority: priority}})

	httpserver.CheckRPCErr(rpcerr, "ModTag")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "ModTag")

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

func getVersions(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "FetchVersions",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "FetchVersions")
	res := resp.Interface().(*config.FetchVersionReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FetchVersions")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addVersion(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	term := req.GetParamInt("term")
	version := req.GetParamInt("version")
	vname := req.GetParamString("vname")
	title := req.GetParamString("title")
	subtitle := req.GetParamString("subtitle")
	downurl := req.GetParamString("downurl")
	desc := req.GetParamString("desc")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "AddVersion",
		&config.VersionRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &config.VersionInfo{Term: term, Version: version,
				Vname: vname, Title: title, Subtitle: subtitle, Downurl: downurl,
				Desc: desc}})

	httpserver.CheckRPCErr(rpcerr, "AddVersion")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddVersion")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func modVersion(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	id := req.GetParamInt("id")
	term := req.GetParamInt("term")
	version := req.GetParamInt("version")
	vname := req.GetParamString("vname")
	title := req.GetParamString("title")
	subtitle := req.GetParamString("subtitle")
	downurl := req.GetParamString("downurl")
	desc := req.GetParamString("desc")
	online := req.GetParamInt("online")
	deleted := req.GetParamInt("deleted")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "ModVersion",
		&config.VersionRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &config.VersionInfo{Term: term, Version: version,
				Vname: vname, Title: title, Subtitle: subtitle, Downurl: downurl,
				Desc: desc, Id: id, Online: online, Deleted: deleted}})

	httpserver.CheckRPCErr(rpcerr, "ModVersion")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "ModVersion")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getUsers(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.UserServerType, req.Uid, "FetchInfos",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Seq: seq, Num: num})

	httpserver.CheckRPCErr(rpcerr, "FetchInfos")
	res := resp.Interface().(*user.FetchInfosReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FetchInfos")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addUser(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	nickname := req.GetParamString("nickname")
	headurl := req.GetParamString("headurl")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.UserServerType, req.Uid, "AddInfo",
		&user.InfoRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &user.Info{Nickname: nickname, Headurl: headurl}})

	httpserver.CheckRPCErr(rpcerr, "AddInfo")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddInfo")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func modUser(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	nickname := req.GetParamString("nickname")
	headurl := req.GetParamString("headurl")
	id := req.GetParamInt("id")
	recommend := req.GetParamInt("recommend")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.UserServerType, req.Uid, "ModInfo",
		&user.ModInfoRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &user.Info{Id: id, Nickname: nickname, Headurl: headurl,
				Recommend: recommend}})

	httpserver.CheckRPCErr(rpcerr, "ModInfo")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "ModInfo")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getShares(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	seq := req.GetParamInt("seq")
	num := req.GetParamInt("num")
	rtype := req.GetParamInt("type")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "FetchShares",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Seq: seq, Num: num, Type: rtype})

	httpserver.CheckRPCErr(rpcerr, "FetchShares")
	res := resp.Interface().(*share.ShareReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FetchShares")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addShareTags(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	id := req.GetParamInt("id")
	tags := req.GetParamWrapIntArray("tags")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "AddShareTags",
		&share.AddTagRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id, Tags: tags})

	httpserver.CheckRPCErr(rpcerr, "AddShareTags")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddShareTags")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func reviewShare(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	id := req.GetParamInt("id")
	reject := req.GetParamIntDef("reject", 0)
	modify := req.GetParamIntDef("modify", 0)
	title := req.GetParamStringDef("title", "")
	smile := req.GetParamIntDef("smile", 0)
	views := req.GetParamIntDef("views", 0)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "ReviewShare",
		&share.ReviewShareRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id, Reject: reject, Modify: modify, Title: title,
			Smile: smile, Views: views})

	httpserver.CheckRPCErr(rpcerr, "ReviewShare")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "ReviewShare")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func searchShare(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	id := req.GetParamInt("id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ShareServerType, req.Uid, "SearchShare",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id})

	httpserver.CheckRPCErr(rpcerr, "SearchShare")
	res := resp.Interface().(*share.ShareReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "SearchShare")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func applyImgUpload(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	size := req.GetParamIntDef("size", 0)
	format := req.GetParamStringDef("format", ".jpg")
	log.Printf("applyImgUpload format:%s %d", format, size)

	filename, auth := ucloud.GenUploadToken(format)
	js, err := simplejson.NewJson([]byte(`{"errno":0}`))
	if err != nil {
		return &util.AppError{Code: httpserver.ErrInner, Msg: err.Error()}
	}
	js.SetPath([]string{"data", "filename"}, filename)
	js.SetPath([]string{"data", "auth"}, auth)
	body, err := js.Encode()
	if err != nil {
		return &util.AppError{Code: httpserver.ErrInner, Msg: err.Error()}
	}

	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func uploadImg(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	r.ParseMultipartForm(10 * 1024 * 1024)
	values := r.MultipartForm.Value["name"]
	log.Printf("values:%v", values)
	files := r.MultipartForm.File["file"]
	var buf []byte
	if len(files) > 0 {
		f, err := files[0].Open()
		if err != nil {
			log.Printf("open file failed:%v", err)
			return &util.AppError{Code: httpserver.ErrInner, Msg: err.Error()}
		}
		buf, err = ioutil.ReadAll(f)
		if err != nil {
			log.Printf("read file failed:%v", err)
			return &util.AppError{Code: httpserver.ErrInner, Msg: err.Error()}
		}
	}
	filename := util.GenUUID() + ".jpg"
	flag := ucloud.PutFile(ucloud.Bucket, filename, buf)
	if !flag {
		log.Printf("ucloud PutFile failed:%s", filename)
		return &util.AppError{Code: httpserver.ErrInner, Msg: "put file failed"}
	}
	js, err := simplejson.NewJson([]byte(`{"errno":0}`))
	if err != nil {
		return &util.AppError{Code: httpserver.ErrInner, Msg: err.Error()}
	}
	js.SetPath([]string{"data", "filename"}, filename)
	body, err := js.Encode()
	if err != nil {
		return &util.AppError{Code: httpserver.ErrInner, Msg: err.Error()}
	}

	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getUserLang(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "FetchUserLang",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid}})

	httpserver.CheckRPCErr(rpcerr, "FetchUserLang")
	res := resp.Interface().(*config.LangReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FetchUserLang")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addUserLang(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	lang := req.GetParamString("lang")
	content := req.GetParamString("content")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "AddUserLang",
		&config.LangRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Info: &config.LangInfo{Lang: lang, Content: content}})

	httpserver.CheckRPCErr(rpcerr, "AddUserLang")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddUserLang")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func delUserLang(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	id := req.GetParamInt("id")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "DelUserLang",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Id: id})

	httpserver.CheckRPCErr(rpcerr, "DelUserLang")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "DelUserLang")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func getLangFollow(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "FetchLangFollow",
		&common.CommRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid}})

	httpserver.CheckRPCErr(rpcerr, "FetchLangFollow")
	res := resp.Interface().(*config.LangFollowReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "FetchLangFollow")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func addLangFollow(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	lid := req.GetParamInt("lid")
	uids := req.GetParamWrapIntArray("uids")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "AddLangFollow",
		&config.LangFollowRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Lid: lid, Uids: uids})

	httpserver.CheckRPCErr(rpcerr, "AddLangFollow")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "AddLangFollow")

	body := httpserver.GenResponseBody(res, false)
	w.Write(body)
	httpserver.ReportSuccResp(r.RequestURI)
	return nil
}

func delLangFollow(w http.ResponseWriter, r *http.Request) (apperr *util.AppError) {
	var req httpserver.Request
	req.InitOss(r)
	ids := req.GetParamWrapIntArray("ids")

	uuid := util.GenUUID()
	resp, rpcerr := httpserver.CallRPC(util.ConfigServerType, req.Uid, "DelLangFollow",
		&config.DelLangFollowRequest{Head: &common.Head{Sid: uuid, Uid: req.Uid},
			Ids: ids})

	httpserver.CheckRPCErr(rpcerr, "DelLangFollow")
	res := resp.Interface().(*common.CommReply)
	httpserver.CheckRPCCode(res.Head.Retcode, "DelLangFollow")

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
	mux.Handle("/mod_tag", httpserver.AppHandler(modTag))
	mux.Handle("/del_tags", httpserver.AppHandler(delTags))
	mux.Handle("/add_version", httpserver.AppHandler(addVersion))
	mux.Handle("/mod_version", httpserver.AppHandler(modVersion))
	mux.Handle("/get_versions", httpserver.AppHandler(getVersions))
	mux.Handle("/get_shares", httpserver.AppHandler(getShares))
	mux.Handle("/get_users", httpserver.AppHandler(getUsers))
	mux.Handle("/add_user", httpserver.AppHandler(addUser))
	mux.Handle("/mod_user", httpserver.AppHandler(modUser))
	mux.Handle("/review_share", httpserver.AppHandler(reviewShare))
	mux.Handle("/search_share", httpserver.AppHandler(searchShare))
	mux.Handle("/add_share_tags", httpserver.AppHandler(addShareTags))
	mux.Handle("/apply_img_upload", httpserver.AppHandler(applyImgUpload))
	mux.Handle("/upload_img", httpserver.AppHandler(uploadImg))
	mux.Handle("/get_user_lang", httpserver.AppHandler(getUserLang))
	mux.Handle("/add_user_lang", httpserver.AppHandler(addUserLang))
	mux.Handle("/del_user_lang", httpserver.AppHandler(delUserLang))
	mux.Handle("/get_user_lang_follow", httpserver.AppHandler(getLangFollow))
	mux.Handle("/add_user_lang_follow", httpserver.AppHandler(addLangFollow))
	mux.Handle("/del_user_lang_follow", httpserver.AppHandler(delLangFollow))
	mux.Handle("/", http.FileServer(http.Dir("/data/laughing/oss")))
	return mux
}
