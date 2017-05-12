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
	return nil
}
