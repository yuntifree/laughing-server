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
	res, err := db.Exec("INSERT INTO media(uid, title, img, dst, abstract, origin, width, height, ctime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW())",
		in.Head.Uid, in.Title, in.Img, in.Dst, in.Desc, in.Origin,
		in.Width, in.Height)
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

	var allowshare int64
	if in.Origin == 0 {
		allowshare = 1
	}
	res, err = db.Exec("INSERT INTO shares(uid, mid, allowshare, ctime) VALUES (?, ?, ?, NOW())",
		in.Head.Uid, mid, allowshare)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	if err != nil {
		return
	}
	_, err = db.Exec("UPDATE users SET videos = videos + 1 WHERE uid = ?",
		in.Head.Uid)
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
	query := "SELECT s.id, s.uid, u.headurl, u.nickname, m.img, m.views, m.title, m.abstract, m.width, m.height, m.id FROM shares s, users u, media m WHERE s.uid = u.uid AND s.mid = m.id "
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
		var mid int64
		err := rows.Scan(&info.Id, &info.Uid, &info.Headurl, &info.Nickname,
			&info.Img, &info.Views, &info.Title, &info.Desc, &info.Width,
			&info.Height, &mid)
		if err != nil {
			log.Printf("getMyShare scan failed:%v", err)
			continue
		}
		nextseq = info.Id
		info.Tags = getMediaTags(db, mid)
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
		var mid int64
		err := rows.Scan(&info.Id, &info.Uid, &info.Headurl, &info.Nickname,
			&info.Img, &info.Views, &info.Title, &info.Desc, &info.Width,
			&info.Height, &mid)
		if err != nil {
			log.Printf("getShare scan failed:%v", err)
			continue
		}
		nextseq = info.Id
		info.Tags = getMediaTags(db, mid)
		infos = append(infos, &info)
	}
	return
}

func genCommentQuery(id, seq, num int64) string {
	query := fmt.Sprintf("SELECT c.id, c.uid, c.content, UNIX_TIMESTAMP(c.ctime), u.headurl, u.nickname FROM comments c, users u WHERE c.uid = u.uid AND c.sid = %d ", id)
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

func getMediaTags(db *sql.DB, id int64) string {
	rows, err := db.Query("SELECT t.content FROM media_tags m, tags t WHERE m.tid = t.id AND m.mid = ?", id)
	if err != nil {
		return ""
	}
	defer rows.Close()
	var tags string
	for rows.Next() {
		var content string
		err := rows.Scan(&content)
		if err != nil {
			log.Printf("getMediaTags scan failed:%v", err)
			continue
		}
		tags += content + " "
	}
	return tags
}

const (
	inner     = 0
	facebook  = 1
	instagram = 2
	musically = 3
)

func getShareNick(db *sql.DB, sid int64) string {
	var nick string
	err := db.QueryRow("SELECT u.nickname FROM users u, shares s WHERE s.uid = u.uid AND s.id = ?", sid).Scan(&nick)
	if err != nil {
		log.Printf("getShareNick failed:%v", err)
	}
	return nick
}

func genShareTitle(db *sql.DB, sid, origin, orisid int64, nickname string) string {
	if orisid == 0 { //not reshare
		switch origin {
		case facebook:
			return nickname + " share from Facebook"
		case instagram:
			return nickname + " share from Instagram"
		case musically:
			return nickname + " share from Musically"
		default:
			return nickname + " Uploaded"
		}
	}
	nick := getShareNick(db, orisid)
	return nickname + " share from " + nick
}

func genShareDesc(minutes int64) string {
	var desc string
	if minutes < 60 {
		desc += fmt.Sprintf(" %d mins ago", minutes)
	} else if minutes < 24*60 {
		desc += fmt.Sprintf(" %d hrs ago", minutes/60)
	} else {
		desc += fmt.Sprintf(" %d days ago", minutes/(60*24))
	}
	return desc
}

func getShareDetail(db *sql.DB, id int64) (info share.ShareDetail, err error) {
	var mid, sid, diff int64
	var record share.ShareRecord
	err = db.QueryRow("SELECT s.reshare, s.comments, s.allowshare, m.img, m.dst, m.title, m.views, m.id, s.sid, u.uid, u.headurl, u.nickname, TIMESTAMPDIFF(MINUTE, s.ctime, NOW()), m.origin FROM shares s, media m, users u WHERE s.mid = m.id AND s.uid = u.uid AND s.id = ?", id).
		Scan(&info.Reshare, &info.Comments, &info.Allowshare, &info.Img, &info.Dst,
			&info.Title, &info.Views, &mid, &sid, &record.Uid, &record.Headurl,
			&record.Nickname, &diff, &record.Origin)
	if err != nil {
		return
	}
	info.Tag = getMediaTags(db, mid)
	record.Desc = genShareDesc(diff)
	record.Title = genShareTitle(db, id, record.Origin, sid, record.Nickname)
	info.Record = &record
	return
}
