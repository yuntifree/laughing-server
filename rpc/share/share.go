package main

import (
	"database/sql"
	"errors"
	"fmt"
	"laughing-server/proto/share"
	"laughing-server/util"
	"log"
	"time"
)

const (
	recommendNum = 10
	interval     = 600
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
	res, err := db.Exec("INSERT INTO media(uid, title, img, dst, abstract, origin, width, height, thumbnail, src, cdn, ctime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())",
		in.Head.Uid, in.Title, in.Img, in.Dst, in.Desc, in.Origin,
		in.Width, in.Height, in.Thumbnail, in.Src, in.Cdn)
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

	res, err = db.Exec("INSERT INTO shares(uid, mid, ctime) VALUES (?, ?, NOW())",
		in.Head.Uid, mid)
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
	var mid, owner int64
	err = db.QueryRow("SELECT m.id, m.uid FROM shares s, media m WHERE s.mid = m.id AND s.id = ?", sid).
		Scan(&mid, &owner)
	if err != nil {
		return
	}
	if uid == owner {
		err = errors.New("can't reshare your own media")
		return
	}

	res, err := db.Exec("INSERT INTO shares (uid, mid, sid, ctime) VALUES (?, ?, ?,  NOW())",
		uid, mid, sid)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	if err != nil {
		return
	}
	_, err = db.Exec("UPDATE shares SET reshare = reshare + 1 WHERE id = ?", sid)
	return
}

func addComment(db *sql.DB, uid, sid int64, content string) (id int64, err error) {
	res, err := db.Exec("INSERT INTO comments(uid, sid, content, ctime) VALUES (?, ?, ?, NOW())", uid, sid, content)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	_, err = db.Exec("UPDATE shares SET comments = comments + 1 WHERE id = ?", sid)
	return
}

func genShareTagQuery(seq, num, id int64) string {
	query := "SELECT s.id, s.uid, u.headurl, u.nickname, m.img, m.views, m.title, m.abstract, m.width, m.height, m.id FROM shares s, users u, media m, media_tags t WHERE  s.mid = t.mid AND s.uid = u.uid AND s.mid = m.id AND s.deleted = 0 "
	query += fmt.Sprintf(" AND t.tid = %d", id)
	if seq != 0 {
		query += fmt.Sprintf(" AND s.id < %d ", seq)
	}
	query += fmt.Sprintf(" ORDER BY s.id DESC LIMIT %d", num)
	return query
}

func genShareQuery(uid, seq, num int64) string {
	query := "SELECT s.id, s.uid, u.headurl, u.nickname, m.img, m.views, m.title, m.abstract, m.width, m.height, m.id FROM shares s, users u, media m WHERE s.uid = u.uid AND s.mid = m.id AND s.deleted = 0 "
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
		infos = append(infos, &info)
	}
	return
}

func getShares(db *sql.DB, seq, num, id int64) (infos []*share.ShareInfo, nextseq int64) {
	var query string
	if id != 0 {
		query = genShareTagQuery(seq, num, id)
	} else {
		query = genShareQuery(0, seq, num)
	}
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
		infos = append(infos, &info)
	}
	return
}

type tagCache struct {
	expired int64
	infos   []*share.TagInfo
}

var tc tagCache

func getRecommendTag(db *sql.DB) *share.TagInfo {
	cnt := len(tc.infos)
	if cnt > 0 && tc.expired > time.Now().Unix() {
		idx := int(util.Rand()) % cnt
		return tc.infos[idx]
	}
	rows, err := db.Query("SELECT id, content, img FROM tags WHERE recommend = 1 AND deleted = 0")
	if err != nil {
		log.Printf("getRecommendTag failed:%v", err)
	}
	var tags []*share.TagInfo
	defer rows.Close()
	for rows.Next() {
		var info share.TagInfo
		err = rows.Scan(&info.Id, &info.Content, &info.Img)
		if err != nil {
			continue
		}
		tags = append(tags, &info)
	}
	tc.infos = tags
	tc.expired = time.Now().Unix() + interval

	cnt = len(tc.infos)
	if cnt > 0 {
		idx := int(util.Rand()) % cnt
		return tc.infos[idx]
	}
	return nil
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

func getMediaTags(db *sql.DB, id int64) []*share.TagInfo {
	rows, err := db.Query("SELECT t.id, t.content FROM media_tags m, tags t WHERE m.tid = t.id AND m.mid = ?", id)
	if err != nil {
		log.Printf("getMediaTags query failed:%v", err)
		return nil
	}
	var infos []*share.TagInfo
	defer rows.Close()
	for rows.Next() {
		var info share.TagInfo
		err := rows.Scan(&info.Id, &info.Content)
		if err != nil {
			log.Printf("getMediaTags scan failed:%v", err)
			continue
		}
		infos = append(infos, &info)
	}
	return infos
}

const (
	inner     = 0
	facebook  = 1
	instagram = 2
	musically = 3
)

func getShareOriNick(db *sql.DB, mid int64) string {
	var nick string
	err := db.QueryRow("SELECT u.nickname FROM users u, media m WHERE m.uid = u.uid AND m.id = ?", mid).Scan(&nick)
	if err != nil {
		log.Printf("getShareNick failed:%v", err)
	}
	return nick
}

func getShareOriTitle(nickname string, origin int64) string {
	prefix := "<b>" + nickname + "</b>"
	switch origin {
	case facebook:
		return prefix + " Share from Facebook"
	case instagram:
		return prefix + " Share from Instagram"
	case musically:
		return prefix + " Share from Musically"
	default:
		return prefix + " Uploaded"
	}
}

func getReshareTitle(db *sql.DB, nickname string, mid int64) string {
	nick := getShareOriNick(db, mid)
	return "<b>" + nickname + "</b>" + " Share from <b>" + nick + "</b>"
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

func hasShare(db *sql.DB, uid, mid int64) int64 {
	var cnt int64
	err := db.QueryRow("SELECT COUNT(id) FROM shares WHERE uid = ? AND mid = ?", uid, mid).Scan(&cnt)
	if err != nil {
		return 0
	}
	if cnt > 0 {
		return 1
	}
	return 0
}

func getShareDetail(db *sql.DB, uid, id int64) (info share.ShareDetail, err error) {
	var mid, sid, diff int64
	var record share.ShareRecord
	err = db.QueryRow("SELECT s.reshare, s.comments, m.img, m.dst, m.title, m.views, m.id, m.width, m.height, s.sid, u.uid, u.headurl, u.nickname, TIMESTAMPDIFF(MINUTE, s.ctime, NOW()), m.origin FROM shares s, media m, users u WHERE s.mid = m.id AND s.uid = u.uid AND s.id = ?", id).
		Scan(&info.Reshare, &info.Comments, &info.Img, &info.Dst,
			&info.Title, &info.Views, &mid, &info.Width, &info.Height,
			&sid, &record.Uid, &record.Headurl,
			&record.Nickname, &diff, &record.Origin)
	if err != nil {
		return
	}
	info.Tags = getMediaTags(db, mid)
	record.Desc = genShareDesc(diff)
	if sid == 0 {
		record.Title = getShareOriTitle(record.Nickname, record.Origin)
	} else {
		record.Title = getReshareTitle(db, record.Nickname, mid)
	}
	info.Record = &record
	info.Hasshare = hasShare(db, uid, mid)
	info.Id = id
	return
}

func unshare(db *sql.DB, uid, sid int64) error {
	res, err := db.Exec("UPDATE shares SET deleted = 1 WHERE uid = ? AND id = ?",
		uid, sid)
	if err != nil {
		return err
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if cnt > 0 {
		_, err = db.Exec("UPDATE users SET videos = IF(videos > 0, videos-1, 0) WHERE uid = ?", uid)
		if err != nil {
			return err
		}
	}
	return nil
}

func getShareIds(db *sql.DB, seq, num, tag int64) (ids []int64, nextseq, nexttag int64) {
	var query string
	if tag == 0 {
		query = "SELECT s.id FROM shares s, media m  WHERE s.mid = m.id "
	} else {
		query = fmt.Sprintf("SELECT s.id FROM shares s, media m, media_tags t WHERE s.mid = m.id AND m.id = t.mid AND t.tid = %d", tag)
	}
	if seq != 0 {
		query += fmt.Sprintf(" AND s.id < %d ", seq)
	}
	query += fmt.Sprintf(" ORDER BY s.id DESC LIMIT %d", num)
	log.Printf("getShareIds query:%s", query)
	rows, err := db.Query(query)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			log.Printf("getShareIds scan failed:%v", err)
			continue
		}
		nextseq = id
		ids = append(ids, id)
	}
	if len(ids) < int(num) && tag != recommendTag {
		rids, next := getRecommendIds(db, 0, num-int64(len(ids)))
		ids = append(ids, rids...)
		nexttag = recommendTag
		nextseq = next
	} else {
		nexttag = tag
	}
	return
}

func getRecommendIds(db *sql.DB, seq, num int64) (ids []int64, nextseq int64) {
	query := fmt.Sprintf("SELECT s.id FROM shares s, media m, media_tags t WHERE s.mid = m.id AND m.id = t.mid AND t.tid = %d", recommendTag)
	if seq != 0 {
		query += fmt.Sprintf(" AND s.id < %d ", seq)
	}
	query += fmt.Sprintf(" ORDER BY s.id DESC LIMIT %d", num)
	rows, err := db.Query(query)
	if err != nil {
		return
	}

	defer rows.Close()
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			log.Printf("getShareIds scan failed:%v", err)
			continue
		}
		nextseq = id
		ids = append(ids, id)
	}
	return
}

func getRecommendShares(db *sql.DB, uid, tag int64) (infos []*share.ShareDetail, err error) {
	var ids []int64
	if tag != 0 {
		ids, _, _ = getShareIds(db, 0, recommendNum, tag)
	} else {
		ids, _ = getRecommendIds(db, 0, recommendNum)
	}
	for _, v := range ids {
		info, err := getShareDetail(db, uid, v)
		if err != nil {
			continue
		}
		infos = append(infos, &info)
	}
	return infos, nil
}
