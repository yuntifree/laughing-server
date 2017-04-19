package main

import (
	"database/sql"
	"laughing-server/util"
	"log"
)

func fblogin(db *sql.DB, fbid, fbtoken string) (uid int64, token, headurl,
	nickname string, err error) {
	err = db.QueryRow("SELECT uid, headurl, nickname FROM users WHERE fb_id = ?", fbid).Scan(&uid, headurl, nickname)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("fblogin query failed:%v", err)
	}
	if uid != 0 {
		token = util.GenSalt()
		_, err = db.Exec("UPDATE users SET token = ?", token)
		if err != nil {
			log.Printf("fblogin update user token  failed:%v", err)
			return
		}
		return
	}
	token = util.GenSalt()
	res, err := db.Exec("INSERT INTO users(token, fb_id, fb_token, ctime) VALUES (?, ?, ?, NOW())",
		token, fbid, fbtoken)
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
