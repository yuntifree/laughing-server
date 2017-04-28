package main

import (
	"database/sql"
	"errors"
	"laughing-server/proto/modify"
)

const (
	mediaView = 1
)

func reportClick(db *sql.DB, in *modify.ClickRequest) (err error) {
	_, err = db.Exec("INSERT INTO click_record(type, uid, cid, imei, ctime) VALUES (?, ?, ?, ?, NOW())", in.Type, in.Head.Uid, in.Id, in.Imei)
	if err != nil {
		return
	}
	switch in.Type {
	case mediaView:
		var mid int64
		err = db.QueryRow("SELECT mid FROM shares WHERE id = ?", in.Id).Scan(&mid)
		if err != nil {
			return
		}
		_, err = db.Exec("UPDATE media SET views = views + 1 WHERE id = ?", in.Id)
	default:
		err = errors.New("illegal report type")
	}
	return
}

func addReport(db *sql.DB, uid, sid int64) error {
	res, err := db.Exec("INSERT IGNORE INTO report(uid, sid, ctime) VALUES (?, ?, NOW())")
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
