package main

import (
	"database/sql"
	"errors"
	"laughing-server/proto/modify"
	"log"
)

const (
	mediaView = 1
)

func reportClick(db *sql.DB, in *modify.ClickRequest) (err error) {
	_, err = db.Exec("INSERT INTO click_record(type, uid, cid, imei, ctime) VALUES (?, ?, ?, ?, NOW())", in.Type, in.Head.Uid, in.Id, in.Imei)
	if err != nil {
		log.Printf("reportClick record failed:%d %d %v", in.Head.Uid, in.Id, err)
		return
	}
	switch in.Type {
	case mediaView:
		var mid int64
		err = db.QueryRow("SELECT mid FROM shares WHERE id = ?", in.Id).Scan(&mid)
		if err != nil {
			log.Printf("reportClick get mid failed:%d %v", in.Id, err)
			return
		}
		log.Printf("mid:%d", mid)
		_, err = db.Exec("UPDATE media SET views = views + 1 WHERE id = ?", mid)
		if err != nil {
			log.Printf("reportClick update views failed:%d %v", in.Id, err)
		}
	default:
		err = errors.New("illegal report type")
	}
	return
}

func addReport(db *sql.DB, uid, sid int64) error {
	res, err := db.Exec("INSERT IGNORE INTO report(uid, sid, ctime) VALUES (?, ?, NOW())", uid, sid)
	if err != nil {
		return err
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if cnt > 0 {
		_, err = db.Exec("UPDATE shares SET report = report + 1 WHERE id = ?", sid)
		if err != nil {
			return err
		}
	}
	var admin int64
	err = db.QueryRow("SELECT admin FROM user WHERE uid = ?", uid).Scan(&admin)
	if err != nil {
		log.Printf("addReport query admin failed:%v", err)
		return nil
	}
	if admin > 0 {
		_, err := db.Exec("UPDATE shares SET deleted = 1 WHERE id = ?", sid)
		if err != nil {
			log.Printf("addReport delete share failed:%d %v", sid, err)
			return nil
		}
		var euid int64
		err = db.QueryRow("SELECT uid FROM shares WHERE id = ?", sid).Scan(&euid)
		if err != nil {
			log.Printf("addReport get share owner failed:%d %v", sid, err)
			return nil
		}
		_, err = db.Exec("UPDATE users SET videos = IF(videos > 1, videos-1, 0) WHERE uid = ?", euid)
		if err != nil {
			log.Printf("addReport update user videos failed:%d %v", euid, err)
			return nil
		}
	}
	return nil
}
