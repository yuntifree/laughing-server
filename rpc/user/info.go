package main

import (
	"database/sql"
	"laughing-server/proto/user"
)

func getInfo(db *sql.DB, tuid int64) (info user.Info, err error) {
	err = db.QueryRow("SELECT headurl, nickname, videos, fan_cnt, follow_cnt FROM users WHERE uid = ? AND deleted = 0", tuid).
		Scan(&info.Headurl, &info.Nickname, &info.Videos, &info.Followers, &info.Following)
	return
}

func modInfo(db *sql.DB, uid int64, headurl, nickname string) error {
	_, err := db.Exec("UPDATE users SET headurl = ?, nickname = ? WHERE uid = ?",
		headurl, nickname, uid)
	return err
}
