package main

import (
	"database/sql"
	fb "laughing-server/facebook"
	"laughing-server/proto/verify"
	"laughing-server/util"
	"log"
)

func fblogin(db *sql.DB, in *verify.FbLoginRequest) (uid int64, token, headurl string, err error) {
	err = db.QueryRow("SELECT uid, headurl FROM users WHERE fb_id = ?", in.Fbid).
		Scan(&uid, &headurl)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("fblogin query failed:%v", err)
	}
	if uid != 0 {
		token = util.GenSalt()
		_, err = db.Exec("UPDATE users SET token = ?, imei = ?, model = ?, language = ?, version = ?, os = ?, nickname = ? WHERE uid = ?",
			token, in.Imei, in.Model, in.Language, in.Version, in.Os, in.Nickname,
			uid)
		if err != nil {
			log.Printf("fblogin update user token  failed:%v", err)
			return
		}
		return
	}
	headurl = fb.GenHeadurl(in.Fbid)
	token = util.GenSalt()
	res, err := db.Exec("INSERT INTO users(token, fb_id, fb_token, imei, model, language, version, os, nickname, headurl, ctime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())",
		token, in.Fbid, in.Fbtoken, in.Imei, in.Model, in.Language,
		in.Version, in.Os, in.Nickname, headurl)
	if err != nil {
		log.Printf("fblogin insert fb info failed:%v", err)
		return
	}
	uid, err = res.LastInsertId()
	if err != nil {
		log.Printf("fblogin get last insert id failed:%v", err)
		return
	}
	return
}

func logout(db *sql.DB, uid int64) {
	_, err := db.Exec("UPDATE users SET token = '' WHERE uid = ?", uid)
	if err != nil {
		log.Printf("logout failed:%v", err)
	}
}
