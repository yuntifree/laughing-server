package main

import (
	"database/sql"
	"laughing-server/proto/user"
)

func getInfo(db *sql.DB, uid, tuid int64) (info user.Info, err error) {
	err = db.QueryRow("SELECT headurl, nickname, videos, fan_cnt, follow_cnt FROM users WHERE uid = ? AND deleted = 0", tuid).
		Scan(&info.Headurl, &info.Nickname, &info.Videos, &info.Followers, &info.Following)

	var cnt int64
	db.QueryRow("SELECT COUNT(id) FROM fan WHERE uid = ? AND tuid = ? AND deleted = 0", tuid, uid).Scan(&cnt)
	if cnt > 0 {
		info.Hasfollow = 1
	}
	return
}

func modInfo(db *sql.DB, uid int64, headurl, nickname string) error {
	_, err := db.Exec("UPDATE users SET headurl = ?, nickname = ? WHERE uid = ?",
		headurl, nickname, uid)
	return err
}
