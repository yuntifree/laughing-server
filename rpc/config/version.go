package main

import (
	"database/sql"
	"laughing-server/proto/config"
	"log"
)

func checkUpdate(db *sql.DB, term, version int64) (vname, desc, title, subtitle, downurl string) {
	err := db.QueryRow("SELECT vname, description, title, subtitle, downurl FROM app_version WHERE term = ? AND version > ? ORDER BY  version DESC LIMIT 1",
		term, version).Scan(&vname, &desc, &title, &subtitle, &downurl)
	if err != nil {
		log.Printf("checkUpdate query failed:%v", err)
	}
	return
}

func fetchVersions(db *sql.DB, seq, num int64) []*config.VersionInfo {
	rows, err := db.Query("SELECT id, term, version, vname, title, subtitle, description, downurl FROM app_version WHERE deleted = 0 ORDER BY ID DESC LIMIT ?, ?",
		seq, num)
	if err != nil {
		log.Printf("fetchVersions query failed:%v", err)
		return nil
	}

	var infos []*config.VersionInfo
	defer rows.Close()
	for rows.Next() {
		var info config.VersionInfo
		err := rows.Scan(&info.Id, &info.Term, &info.Version, &info.Vname,
			&info.Title, &info.Subtitle, &info.Desc, &info.Downurl)
		if err != nil {
			log.Printf("fetchVersions scan failed:%v", err)
			continue
		}
		infos = append(infos, &info)
	}
	return infos
}

func getTotalVersions(db *sql.DB) int64 {
	var cnt int64
	err := db.QueryRow("SELECT COUNT(id) FROM app_version WHERE deleted = 0").Scan(&cnt)
	if err != nil {
		log.Printf("getTotalVersions query failed:%v", err)
	}
	return cnt
}

func addVersion(db *sql.DB, info *config.VersionInfo) (id int64, err error) {
	res, err := db.Exec("INSERT INTO app_version(term, version, vname, title, subtitle, downurl, description, ctime) VALUES (?, ?, ?, ?, ?, ?, ?, NOW())",
		info.Term, info.Version, info.Vname, info.Title, info.Subtitle,
		info.Downurl, info.Desc)
	if err != nil {
		log.Printf("addVersion insert failed:%v", err)
		return 0, err
	}
	id, err = res.LastInsertId()
	if err != nil {
		log.Printf("addVersion get insert id failed:%v", err)
	}
	return
}
