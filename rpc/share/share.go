package main

import (
	"database/sql"
	"errors"
	"fmt"
	"laughing-server/proto/share"
	"laughing-server/ucloud"
	"laughing-server/util"
	"log"
)

const (
	recommendNum = 10
	interval     = 600
)

func addMediaTags(db *sql.DB, mid int64, tags []int64) {
	for i := 0; i < len(tags); i++ {
		_, err := db.Exec("INSERT INTO media_tags(mid, tid, ctime) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE deleted = 0", mid, tags[i])
		if err != nil {
			log.Printf("addMediaTags failed:%v", err)
			continue
		}
	}
}

func addShare(db *sql.DB, in *share.ShareRequest) (id int64, err error) {
	if in.Origin == 0 && in.Dst != "" {
		in.Cdn = in.Dst
	}
	res, err := db.Exec("INSERT INTO media(uid, title, img, dst, abstract, origin, width, height, thumbnail, src, cdn, views, smile, ctime) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())",
		in.Head.Uid, in.Title, in.Img, in.Dst, in.Desc, in.Origin,
		in.Width, in.Height, in.Thumbnail, in.Src, in.Cdn, in.Views,
		in.Smile)
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

	res, err = db.Exec("INSERT INTO shares(uid, mid, review, ctime) VALUES (?, ?, ?, NOW())",
		in.Head.Uid, mid, in.Review)
	if err != nil {
		return
	}
	id, err = res.LastInsertId()
	if err != nil {
		return
	}
	_, err = db.Exec("UPDATE users SET videos = videos + 1 WHERE uid = ?",
		in.Head.Uid)
	if in.Origin != 0 && in.Img == "" {
		err := util.PubData(w, "shares", map[string]interface{}{"sid": id})
		if err != nil {
			log.Printf("addShare PubData failed:%d %v", id, err)
		}
	}
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

	res, err := db.Exec("INSERT INTO shares (uid, mid, sid, review, ctime) VALUES (?, ?, ?, 1, NOW()) ON DUPLICATE KEY UPDATE deleted = 0",
		uid, mid, sid)
	if err != nil {
		return
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		log.Printf("reshare get rows affected failed:%d %d %v",
			uid, sid, err)
		return
	}
	if cnt == 0 {
		return
	}
	id, err = res.LastInsertId()
	if err != nil {
		log.Printf("reshare get last insert id failed:%d %d %v",
			uid, sid, err)
		return
	}
	if id == 0 {
		return
	}
	_, err = db.Exec("UPDATE shares SET reshare = reshare + 1 WHERE id = ?", sid)
	_, err = db.Exec("UPDATE users SET videos = videos + 1 WHERE uid = ?", uid)
	return
}

func addComment(db *sql.DB, uid, sid int64, content string) (id int64, err error) {
	res, err := db.Exec("INSERT INTO comments(uid, sid, content, ctime) VALUES (?, ?, ?, NOW())", uid, sid, content)
	if err != nil {
		log.Printf("addComment insert comments failed:%v", err)
		return
	}
	id, err = res.LastInsertId()
	if err != nil {
		log.Printf("addComment get last insert id failed:%v", err)
	}
	_, err = db.Exec("UPDATE shares SET comments = comments + 1 WHERE id = ?", sid)
	if err != nil {
		log.Printf("addComment update share comments failed:%v", err)
	}
	return
}

func genShareTagQuery(uid, seq, num, id int64) string {
	query := "SELECT s.id, s.uid, u.headurl, u.nickname, m.img, m.views, m.title, m.abstract, m.width, m.height, m.id, m.smile, s.review FROM shares s, users u, media m, media_tags t  WHERE  s.mid = t.mid AND s.uid = u.uid AND s.mid = m.id AND s.deleted = 0 AND m.smile != 0 AND s.sid = 0 "
	query += fmt.Sprintf(" AND t.tid = %d", id)
	if seq != 0 {
		query += fmt.Sprintf(" AND s.id < %d ", seq)
	}
	query += fmt.Sprintf(" ORDER BY s.id DESC LIMIT %d", num)
	return query
}

func genShareQuery(uid, tuid, seq, num int64) string {
	query := "SELECT s.id, s.uid, u.headurl, u.nickname, m.img, m.views, m.title, m.abstract, m.width, m.height, m.id, m.smile, s.review FROM shares s, users u, media m WHERE s.uid = u.uid AND s.mid = m.id AND s.deleted = 0 "
	if uid != 0 && tuid == 0 {
		query += fmt.Sprintf(" AND (s.uid IN (SELECT tuid FROM follower WHERE uid = %d AND deleted = 0) OR s.uid = %d) ", uid, uid)
	} else if tuid == 0 {
		query += " AND u.recommend = 1 "
	}
	if uid != tuid || (uid == 0 && tuid == 0) {
		query += fmt.Sprintf(" AND m.smile != 0 ")
	}
	if seq != 0 {
		query += fmt.Sprintf(" AND s.id < %d ", seq)
	}
	if tuid != 0 {
		query += fmt.Sprintf(" AND s.uid = %d ", tuid)
	}
	query += fmt.Sprintf(" ORDER BY s.id DESC LIMIT %d", num)
	return query
}

func getUserShares(db *sql.DB, uid, tuid, seq, num int64) (infos []*share.ShareInfo, nextseq int64) {
	query := genShareQuery(uid, tuid, seq, num)
	log.Printf("query:%s", query)
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
			&info.Height, &mid, &info.Smile, &info.Review)
		if err != nil {
			log.Printf("getMyShare scan failed:%v", err)
			continue
		}
		nextseq = info.Id
		info.Headurl = ucloud.GenHeadThumb(info.Headurl)
		info.Img = ucloud.GenImgThumb(info.Img)
		infos = append(infos, &info)
	}
	return
}

func getShares(db *sql.DB, uid, seq, num, id int64) (infos []*share.ShareInfo, nextseq int64) {
	var query string
	if id != 0 {
		query = genShareTagQuery(uid, seq, num, id)
	} else {
		query = genShareQuery(uid, 0, seq, num)
	}
	rows, err := db.Query(query)
	log.Printf("query:%s", query)
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
			&info.Height, &mid, &info.Smile, &info.Review)
		if err != nil {
			log.Printf("getShare scan failed:%v", err)
			continue
		}
		nextseq = info.Id
		info.Img = ucloud.GenImgThumb(info.Img)
		info.Headurl = ucloud.GenHeadThumb(info.Headurl)
		infos = append(infos, &info)
	}
	return
}

func getRecommendTag(db *sql.DB) *share.TagInfo {
	var info share.TagInfo
	err := db.QueryRow("SELECT id, content, img FROM tags WHERE recommend = 1 AND deleted = 0 LIMIT 1").Scan(&info.Id, &info.Content, &info.Img)
	if err != nil {
		log.Printf("getRecommendTag failed:%v", err)
		return nil
	}
	info.Img = ucloud.GenImgThumb(info.Img)
	return &info
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
		info.Headurl = ucloud.GenHeadurl(info.Headurl)
		infos = append(infos, &info)
	}
	return
}

func getMediaTags(db *sql.DB, id int64) []*share.TagInfo {
	rows, err := db.Query("SELECT t.id, t.content FROM media_tags m, tags t WHERE m.tid = t.id AND m.deleted = 0 AND m.mid = ?", id)
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

func getShareOriInfo(db *sql.DB, mid int64) (uid int64, nickname string) {
	err := db.QueryRow("SELECT u.uid, u.nickname FROM users u, media m WHERE m.uid = u.uid AND m.id = ?", mid).Scan(&uid, &nickname)
	if err != nil {
		log.Printf("getShareNick failed:%v", err)
	}
	return
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
	err := db.QueryRow("SELECT COUNT(id) FROM shares WHERE uid = ? AND mid = ? AND deleted = 0", uid, mid).Scan(&cnt)
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
	err = db.QueryRow("SELECT s.reshare, s.comments, m.img, m.cdn, m.title, m.views, m.id, m.width, m.height, m.unshare, s.sid, u.uid, u.headurl, u.nickname, TIMESTAMPDIFF(MINUTE, s.ctime, NOW()), m.origin, m.smile, m.src FROM shares s, media m, users u WHERE s.mid = m.id AND s.uid = u.uid AND s.id = ?", id).
		Scan(&info.Reshare, &info.Comments, &info.Img, &info.Dst,
			&info.Title, &info.Views, &mid, &info.Width, &info.Height,
			&info.Unshare, &sid, &record.Uid, &record.Headurl,
			&record.Nickname, &diff, &record.Origin, &info.Smile, &record.Src)
	if err != nil {
		return
	}
	if info.Unshare > 0 {
		info.Dst = ""
	}
	info.Img = ucloud.GetCdnURL(info.Img)
	info.Dst = ucloud.GetCdnURL(info.Dst)
	info.Tags = getMediaTags(db, mid)
	record.Desc = genShareDesc(diff)
	record.Headurl = ucloud.GenHeadurl(record.Headurl)
	if sid == 0 {
		record.Oriuid = record.Uid
		record.Orinick = record.Nickname
		record.Uid = 0
		record.Nickname = ""
	} else {
		record.Oriuid, record.Orinick = getShareOriInfo(db, mid)
	}
	info.Record = &record
	info.Hasshare = hasShare(db, uid, mid)
	info.Id = id
	return
}

func deleteShareRecord(db *sql.DB, uid, mid int64) bool {
	res, err := db.Exec("UPDATE shares SET deleted = 1 WHERE uid = ? AND mid = ?",
		uid, mid)
	if err != nil {
		log.Printf("deleteShareRecord update failed:%d %d %v",
			uid, mid, err)
		return false
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		log.Printf("deleteShareRecord get affect rows failed:%d %d %v",
			uid, mid, err)
		return false
	}
	if cnt > 0 {
		return true
	}
	return false
}

func decrUserVideo(db *sql.DB, uid int64) bool {
	_, err := db.Exec("UPDATE users SET videos = IF(videos > 0, videos-1, 0) WHERE uid = ?", uid)
	if err != nil {
		log.Printf("decrUserVideo failed:%v", err)
		return false
	}
	return true
}

func unshareMedia(db *sql.DB, mid int64) bool {
	_, err := db.Exec("UPDATE media SET unshare = 1 WHERE id = ?", mid)
	if err != nil {
		log.Printf("unshare update media failed:%v", err)
		return false
	}
	return true
}

func unshareOwner(db *sql.DB, uid, mid int64) {
	if !deleteShareRecord(db, uid, mid) {
		return
	}
	decrUserVideo(db, uid)
	unshareMedia(db, mid)
	return
}

func unshareOthers(db *sql.DB, uid, mid int64) {
	if !deleteShareRecord(db, uid, mid) {
		return
	}
	decrUserVideo(db, uid)
	var orisid int64
	err := db.QueryRow("SELECT sid FROM shares WHERE uid = ? AND mid = ?",
		uid, mid).Scan(&orisid)
	if err != nil {
		log.Printf("unshareOther query origin sid failed:%d %d %v",
			uid, mid, err)
		return
	}
	if orisid != 0 {
		_, err = db.Exec("UPDATE shares SET reshare = IF(reshare > 0, reshare-1, 0) WHERE id = ?", orisid)
		if err != nil {
			log.Printf("unshare minus origin sid reshare failed:%v", err)
			return
		}
	}
	return
}

func unshare(db *sql.DB, uid, sid int64) error {
	var euid, mid int64
	err := db.QueryRow("SELECT m.id, m.uid FROM media m, shares s WHERE s.mid = m.id AND s.id = ?", sid).Scan(&mid, &euid)
	if err != nil {
		log.Printf("unshare query euid failed:%v", err)
		return err
	}

	if uid == euid {
		unshareOwner(db, uid, mid)
	} else {
		unshareOthers(db, uid, mid)
	}
	return nil
}

func getShareIds(db *sql.DB, seq, num, tag, sid int64) (ids []int64, nextseq, nexttag int64) {
	var query string
	if tag == 0 {
		query = "SELECT s.id FROM shares s, media m  WHERE s.mid = m.id "
	} else {
		query = fmt.Sprintf("SELECT s.id FROM shares s, media m, media_tags t WHERE s.mid = m.id AND m.id = t.mid AND t.tid = %d", tag)
	}
	query += " AND s.sid = 0 AND s.deleted = 0 AND m.smile > 0 "
	if sid != 0 {
		query += fmt.Sprintf(" AND s.id < %d", sid)
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
		rids, next := getRecommendIds(db, 0, num-int64(len(ids)), 0)
		ids = append(ids, rids...)
		nexttag = recommendTag
		nextseq = next
	} else {
		nexttag = tag
	}
	return
}

func getRecommendIds(db *sql.DB, seq, num, sid int64) (ids []int64, nextseq int64) {
	query := fmt.Sprintf("SELECT s.id FROM shares s, media m, media_tags t, tags ts WHERE s.mid = m.id AND m.id = t.mid AND t.tid = ts.id AND ts.hot = 1")
	if sid != 0 {
		query += fmt.Sprintf(" AND s.id < %d", sid)
	}
	if seq != 0 {
		query += fmt.Sprintf(" AND s.id < %d ", seq)
	}
	query += " AND s.sid = 0 AND s.deleted = 0 AND m.smile > 0 "
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

func getRecommendShares(db *sql.DB, uid, tag, sid int64) (infos []*share.ShareDetail, err error) {
	var ids []int64
	if tag != 0 {
		ids, _, _ = getShareIds(db, 0, recommendNum, tag, sid)
	} else {
		ids, _ = getRecommendIds(db, 0, recommendNum, sid)
	}
	for _, v := range ids {
		info, err := getShareDetail(db, uid, v)
		if err != nil {
			continue
		}
		info.Img = info.Img + ucloud.Thumbnail
		infos = append(infos, &info)
	}
	return infos, nil
}

func fetchShares(db *sql.DB, seq, num, rtype int64) []*share.ShareInfo {
	query := "SELECT s.id, s.uid, u.headurl, u.nickname, m.img, m.views, m.title, m.abstract, m.id, m.smile FROM shares s, users u, media m WHERE s.uid = u.uid AND s.mid = m.id AND s.deleted = 0 AND s.sid = 0 "
	query += fmt.Sprintf(" AND review = %d", rtype)
	query += fmt.Sprintf(" ORDER BY s.id DESC LIMIT %d, %d", seq, num)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("fetchShares query failed:%v", err)
		return nil
	}
	var infos []*share.ShareInfo
	defer rows.Close()
	for rows.Next() {
		var info share.ShareInfo
		var mid int64
		err = rows.Scan(&info.Id, &info.Uid, &info.Headurl, &info.Nickname,
			&info.Img, &info.Views, &info.Title, &info.Desc, &mid, &info.Smile)
		if err != nil {
			log.Printf("fetchShares scan failed:%v", err)
			continue
		}
		info.Headurl = ucloud.GenHeadurl(info.Headurl)
		info.Img = ucloud.GetCdnURL(info.Img)
		info.Taginfo = getMediaTags(db, mid)
		infos = append(infos, &info)
	}
	return infos
}

func searchShares(db *sql.DB, sid int64) []*share.ShareInfo {
	query := "SELECT s.id, s.uid, u.headurl, u.nickname, m.img, m.views, m.title, m.abstract, m.id, m.smile FROM shares s, users u, media m WHERE s.uid = u.uid AND s.mid = m.id AND s.deleted = 0 "
	query += fmt.Sprintf(" AND s.id = %d ORDER BY s.id DESC", sid)
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("searchShares query failed:%v", err)
		return nil
	}
	var infos []*share.ShareInfo
	defer rows.Close()
	for rows.Next() {
		var info share.ShareInfo
		var mid int64
		err = rows.Scan(&info.Id, &info.Uid, &info.Headurl, &info.Nickname,
			&info.Img, &info.Views, &info.Title, &info.Desc, &mid, &info.Smile)
		if err != nil {
			log.Printf("searchShares scan failed:%v", err)
			continue
		}
		info.Headurl = ucloud.GenHeadurl(info.Headurl)
		info.Img = ucloud.GetCdnURL(info.Img)
		info.Taginfo = getMediaTags(db, mid)
		infos = append(infos, &info)
	}
	return infos
}

func getStrTags(db *sql.DB, mid int64) string {
	rows, err := db.Query("SELECT t.content FROM media_tags m, tags t WHERE m.tid = t.id AND m.mid = ? AND m.deleted = 0", mid)
	if err != nil {
		log.Printf("getStrTags query failed:%v", err)
		return ""
	}
	var tags string
	defer rows.Close()
	for rows.Next() {
		var content string
		err := rows.Scan(&content)
		if err != nil {
			log.Printf("getStrTags scan failed:%v", err)
			continue
		}
		tags += content + " "
	}
	return tags
}

func getTotalShares(db *sql.DB, rtype int64) int64 {
	var cnt int64
	query := "SELECT COUNT(id) FROM shares WHERE deleted = 0 "
	if rtype != 0 {
		query += fmt.Sprintf(" AND review = %d", rtype)
	}
	err := db.QueryRow(query).Scan(&cnt)
	if err != nil {
		log.Printf("getTotalShares failed:%v", err)
	}
	return cnt
}

func getShareMid(db *sql.DB, sid int64) int64 {
	var id int64
	err := db.QueryRow("SELECT mid FROM shares WHERE id = ?", sid).Scan(&id)
	if err != nil {
		log.Printf("getShareMid failed:%v", err)
	}
	return id
}

func reviewShare(db *sql.DB, in *share.ReviewShareRequest) {
	if in.Reject > 0 {
		_, err := db.Exec("UPDATE shares SET review = 2, deleted = 1 WHERE id = ?", in.Id)
		if err != nil {
			log.Printf("reviewShare update shares failed:%v", err)
		}
		return
	}

	if in.Modify > 0 {
		_, err := db.Exec("UPDATE shares SET review = 1 WHERE id = ?", in.Id)
		if err != nil {
			log.Printf("reviewShare update shares failed:%v", err)
			return
		}
	} else {
		_, err := db.Exec("UPDATE shares SET review = 1  WHERE id = ?", in.Id)
		if err != nil {
			log.Printf("reviewShare update shares failed:%v", err)
			return
		}
	}
	var mid int64
	err := db.QueryRow("SELECT mid FROM shares WHERE id = ?", in.Id).Scan(&mid)
	if err != nil {
		log.Printf("reviewShare get media id failed:%v", err)
		return
	}
	if in.Modify > 0 {
		_, err = db.Exec("UPDATE media SET smile = ?, title = ? WHERE id = ?",
			in.Smile, in.Title, mid)
	} else {
		_, err = db.Exec("UPDATE media SET smile = ? WHERE id = ?",
			in.Smile, mid)
	}
	if err != nil {
		log.Printf("reviewShare set smile failed:%v", err)
		return
	}
	return
}

func addShareTag(db *sql.DB, id int64, tags []int64) {
	mid := getShareMid(db, id)
	_, err := db.Exec("UPDATE media_tags SET deleted = 1 WHERE mid = ?", mid)
	if err != nil {
		log.Printf("addShareTag delete old tag failed:%v", err)
	}
	if mid > 0 {
		addMediaTags(db, mid, tags)
	}
}
