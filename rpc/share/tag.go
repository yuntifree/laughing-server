package main

import (
	"database/sql"
	"fmt"
	"laughing-server/proto/share"
	"laughing-server/ucloud"
	"laughing-server/util"
	"log"
)

func getTags(db *sql.DB) []*share.TagInfo {
	var infos []*share.TagInfo
	rows, err := db.Query("SELECT id, content FROM tags WHERE deleted = 0 AND hot = 1 ORDER BY priority DESC")
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

func fetchTags(db *sql.DB, seq, num int64) []*share.TagInfo {
	var infos []*share.TagInfo
	rows, err := db.Query("SELECT id, content, img, recommend, hot, priority FROM tags WHERE deleted = 0 ORDER BY id DESC LIMIT ?, ?", seq, num)
	if err != nil {
		log.Printf("fetchTags query failed:%v", err)
		return infos
	}

	defer rows.Close()
	for rows.Next() {
		var info share.TagInfo
		err := rows.Scan(&info.Id, &info.Content, &info.Img, &info.Recommend,
			&info.Hot, &info.Priority)
		if err != nil {
			log.Printf("fetchTags scan failed:%v", err)
			continue
		}
		info.Img = ucloud.GenHeadurl(info.Img)
		infos = append(infos, &info)
	}
	return infos
}

func getTotalTags(db *sql.DB) int64 {
	var cnt int64
	err := db.QueryRow("SELECT COUNT(id) FROM tags WHERE deleted = 0").Scan(&cnt)
	if err != nil {
		log.Printf("getTotalTags query failed:%v", err)
	}
	return cnt
}

func addTag(db *sql.DB, info *share.TagInfo) (id int64, err error) {
	res, err := db.Exec("INSERT INTO tags(content, img, recommend, hot, ctime) VALUES (?, ?, ?, ?, NOW())",
		info.Content, info.Img, info.Recommend, info.Hot)
	if err != nil {
		log.Printf("addTag err:%s %v", info.Content, err)
		return 0, err
	}

	id, err = res.LastInsertId()
	if info.Recommend == 1 {
		_, err = db.Exec("UPDATE tags SET recommend = 0 WHERE id != ?", id)
		if err != nil {
			log.Printf("addTag update failed:%v", err)
		}

	}
	return
}

func genDelTagQuery(ids []int64) string {
	query := "UPDATE tags SET deleted = 1 WHERE id IN ("
	for i := 0; i < len(ids); i++ {
		query += fmt.Sprintf("%d,", ids[i])
	}
	query += "0)"
	return query
}

func delTags(db *sql.DB, ids []int64) error {
	query := genDelTagQuery(ids)
	_, err := db.Exec(query)
	if err != nil {
		log.Printf("delTags failed:%s %v", query, err)
	}
	return err
}

func modTag(db *sql.DB, info *share.TagInfo) error {
	img := util.ExtractFilename(info.Img)
	_, err := db.Exec("UPDATE tags SET img = ?, content = ?, recommend = ?, hot = ?, priority = ? WHERE id = ?",
		img, info.Content, info.Recommend, info.Hot, info.Priority, info.Id)
	if err != nil {
		log.Printf("modTag query failed:%v", err)
		return err
	}
	if info.Recommend == 1 {
		_, err := db.Exec("UPDATE tags SET recommend = 0 WHERE id != ?", info.Id)
		if err != nil {
			log.Printf("modTag update recommend failed:%v", err)
			return err
		}
	}
	return err
}
