package main

import (
	"database/sql"
	"errors"
	"fmt"
	"laughing-server/proto/share"
	"log"
)

func addMediaTags(db *sql.DB, mid int64, tags []int64) {
	query := "INSERT INTO media_tags(mid, tid) VALUES "
	for i := 0; i < len(tags); i++ {
		if i == len(tags)-1 {
			query += fmt.Sprintf(" (%d, %d)", mid, tags[i])
		} else {
			query += fmt.Sprintf(" (%d, %d),", mid, tags[i])
		}
	}
	_, err := db.Exec(query)
	log.Printf("addMediaTags query:%s", query)
	if err != nil {
		log.Printf("addMediaTags query failed:%v", err)
	}
}

func addShare(db *sql.DB, in *share.ShareRequest) (id int64, err error) {
	res, err := db.Exec("INSERT INTO media(uid, title, img, dst, abstract, origin, ctime) VALUES (?, ?, ?, ?, ?, ?, NOW())",
		in.Head.Uid, in.Title, in.Img, in.Dst, in.Abstract, in.Origin)
	if err != nil {
		return
	}
	mid, err := res.LastInsertId()
	if err != nil {
		return
	}
	if len(in.Tags) > 0 {
		addMediaTags(db, mid, in.Tags)
	}

	res, err = db.Exec("INSERT INTO shares(uid, mid, allowshare, ctime) VALUES (?, ?, ?, NOW())",
		in.Head.Uid, mid, in.Origin)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	if err != nil {
		return
	}
	return
}

func reshare(db *sql.DB, uid, sid int64) (id int64, err error) {
	var mid, allowshare int64
	err = db.QueryRow("SELECT mid, allowshare FROM shares WHERE id = ?", sid).
		Scan(&mid, &allowshare)
	if err != nil {
		return
	}
	if allowshare != 1 {
		err = errors.New("reshare not allowed")
		return
	}

	res, err := db.Exec("INSERT INTO shares (uid, mid, sid, allowshare, ctime) VALUES (?, ?, ?, 0, NOW())", uid, mid, sid)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	if err != nil {
		return
	}
	return
}

func addComment(db *sql.DB, uid, sid int64, content string) (id int64, err error) {
	res, err := db.Exec("INSERT INTO comments(uid, sid, content, ctime) VALUES (?, ?, ?, NOW())", uid, sid, content)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	return
}
