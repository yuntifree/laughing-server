package main

import (
	"database/sql"
	"laughing-server/proto/user"
	"log"
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

func fetchInfos(db *sql.DB, seq, num int64) []*user.Info {
	rows, err := db.Query("SELECT uid, imei, headurl, nickname, fan_cnt, follow_cnt, videos, ctime FROM users ORDER BY uid DESC LIMIT ?, ?", seq, num)
	if err != nil {
		log.Printf("fetchInfos query failed:%v", err)
		return nil
	}
	var infos []*user.Info
	defer rows.Close()
	for rows.Next() {
		var info user.Info
		err = rows.Scan(&info.Id, &info.Imei, &info.Headurl, &info.Nickname,
			&info.Followers, &info.Following, &info.Videos, &info.Ctime)
		if err != nil {
			log.Printf("fetchInfos scan failed:%v", err)
			continue
		}
		infos = append(infos, &info)
	}
	return infos
}

func addInfo(db *sql.DB, info *user.Info) (id int64, err error) {
	res, err := db.Exec("INSERT INTO users(headurl, nickname, ctime) VALUES (?, ?, NOW())",
		info.Headurl, info.Nickname)
	if err != nil {
		log.Printf("addInfo insert failed:%v", err)
		return
	}
	id, err = res.LastInsertId()
	return
}

func getTotalUsers(db *sql.DB) int64 {
	var cnt int64
	err := db.QueryRow("SELECT COUNT(uid) FROM users WHERE deleted = 0").Scan(&cnt)
	if err != nil {
		log.Printf("getTotalUsers query failed:%v", err)
	}
	return cnt
}
