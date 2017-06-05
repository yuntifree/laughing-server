package main

import (
	"database/sql"
	"laughing-server/proto/config"
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
