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

func genShareQuery(uid, seq, num int64) string {
	query := "SELECT s.id, s.uid, u.headurl, u.nickname, m.img, m.views, m.title, m.abstract FROM shares s, users u, media m WHERE s.uid = u.uid AND s.mid = m.id "
	if seq != 0 {
		query += fmt.Sprintf(" AND s.id < %d ", seq)
	}
	if uid != 0 {
		query += fmt.Sprintf(" AND s.uid = %d ", uid)
	}
	query += fmt.Sprintf(" ORDER BY s.id DESC LIMIT %d", num)
	return query
}

func getMyShares(db *sql.DB, uid, seq, num int64) (infos []*share.ShareInfo, nextseq int64) {
	query := genShareQuery(uid, seq, num)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("getMyShares query failed:%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var info share.ShareInfo
		err := rows.Scan(&info.Id, &info.Uid, &info.Headurl, &info.Nickname,
			&info.Img, &info.Views, &info.Title, &info.Abstract)
		if err != nil {
			log.Printf("getMyShare scan failed:%v", err)
			continue
		}
		nextseq = info.Id
		infos = append(infos, &info)
	}
	return
}

func getShares(db *sql.DB, seq, num int64) (infos []*share.ShareInfo, nextseq int64) {
	query := genShareQuery(0, seq, num)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("getShares query failed:%v", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var info share.ShareInfo
		err := rows.Scan(&info.Id, &info.Uid, &info.Headurl, &info.Nickname,
			&info.Img, &info.Views, &info.Title, &info.Abstract)
		if err != nil {
			log.Printf("getShare scan failed:%v", err)
			continue
		}
		nextseq = info.Id
		infos = append(infos, &info)
	}
	return
}

func genCommentQuery(id, seq, num int64) string {
	query := fmt.Sprintf("SELECT c.id, c.uid, c.content, c.ctime, u.headurl, u.nickname FROM comments c, users u WHERE c.uid = u.uid AND c.sid = %d ", id)
	if seq != 0 {
		query += fmt.Sprintf(" AND c.id < %d ", seq)
	}
	query += fmt.Sprintf(" ORDER BY c.id DESC LIMIT %d", num)
	return query
}

func getShareComments(db *sql.DB, id, seq, num int64) (infos []*share.CommentInfo, nextseq int64) {
	query := genCommentQuery(id, seq, num)
	rows, err := db.Query(query)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var info share.CommentInfo
		err := rows.Scan(&info.Id, &info.Uid, &info.Content, &info.Ctime,
			&info.Headurl, &info.Nickname)
		if err != nil {
			log.Printf("getShareComments scan failed:%v", err)
			continue
		}
		nextseq = info.Id
		infos = append(infos, &info)
	}
	return
}
