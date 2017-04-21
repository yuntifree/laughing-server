package main

import (
	"database/sql"
	"laughing-server/proto/share"
	"log"
)

func getTags(db *sql.DB) []*share.TagInfo {
	var infos []*share.TagInfo
	rows, err := db.Query("SELECT id, content FROM tags WHERE deleted = 0")
	if err != nil {
		log.Printf("getTags query failed:%v", err)
		return infos
	}

	defer rows.Close()
	for rows.Next() {
		var info share.TagInfo
		err := rows.Scan(&info.Id, &info.Content)
		if err != nil {
			log.Printf("getTags scan failed:%v", err)
			continue
		}
		infos = append(infos, &info)
	}
	return infos
}
