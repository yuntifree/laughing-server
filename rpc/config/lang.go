package main

import (
	"database/sql"
	"fmt"
	"laughing-server/proto/config"
	"laughing-server/ucloud"
	"log"
)

func fetchUserLang(db *sql.DB) []*config.LangInfo {
	var infos []*config.LangInfo
	rows, err := db.Query("SELECT id, lang, content FROM user_lang WHERE deleted = 0")
	if err != nil {
		log.Printf("fetchUserLang failed:%v", err)
		return infos
	}
	defer rows.Close()
	for rows.Next() {
		var info config.LangInfo
		err = rows.Scan(&info.Id, &info.Lang, &info.Content)
		if err != nil {
			log.Printf("fetchUserLang scan failed:%v", err)
			continue
		}
		infos = append(infos, &info)
	}
	return infos
}

func addUserLang(db *sql.DB, info *config.LangInfo) (int64, error) {
	res, err := db.Exec("INSERT INTO user_lang(lang, content, ctime) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE content = ?, deleted = 0",
		info.Lang, info.Content, info.Content)
	if err != nil {
		log.Printf("addUserLang insert failed:%v", err)
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		log.Printf("addUserLang get insert id failed:%v", err)
		return 0, err
	}
	return id, nil
}

func delUserLang(db *sql.DB, id int64) error {
	_, err := db.Exec("UPDATE user_lang SET deleted = 1 WHERE id = ?", id)
	return err
}

func fetchLangFollow(db *sql.DB) []*config.LangFollowInfo {
	rows, err := db.Query("SELECT f.id, f.lid, f.uid, l.lang, l.content, u.headurl, u.nickname FROM lang_follower f, user_lang l, users u WHERE f.lid = l.id AND f.uid = u.uid AND f.deleted = 0")
	if err != nil {
		log.Printf("fetchLangFollow query failed:%v", err)
		return nil
	}
	defer rows.Close()
	var infos []*config.LangFollowInfo
	for rows.Next() {
		var info config.LangFollowInfo
		err := rows.Scan(&info.Id, &info.Lid, &info.Uid, &info.Lang,
			&info.Content, &info.Headurl, &info.Nickname)
		if err != nil {
			log.Printf("fetchLangFollow scan failed:%v", err)
			continue
		}
		info.Headurl = ucloud.GenHeadurl(info.Headurl)
		infos = append(infos, &info)
	}
	return infos
}

func recordLangFollow(db *sql.DB, lid, uid int64) {
	_, err := db.Exec("INSERT INTO lang_follower(lid, uid, ctime) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE deleted = 0",
		lid, uid)
	if err != nil {
		log.Printf("recrodLangFollow failed:%d %d %v", lid, uid, err)
	}
}

func addLangFollow(db *sql.DB, lid int64, uids []int64) {
	if len(uids) == 0 {
		return
	}
	for i := 0; i < len(uids); i++ {
		recordLangFollow(db, lid, uids[i])
	}
}

func delLangFollow(db *sql.DB, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	query := "UPDATE lang_follower SET deleted = 1 WHERE id IN ("
	for i := 0; i < len(ids); i++ {
		query += fmt.Sprintf("%d,", ids[i])
	}
	query += ",0)"
	_, err := db.Exec(query)
	return err
}
