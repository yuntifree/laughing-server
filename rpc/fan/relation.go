package main

import (
	"database/sql"
	"fmt"
	"laughing-server/proto/fan"
	"log"
)

const (
	fanType      = 0
	followerType = 1
)

func getRelations(db *sql.DB, uid, rtype, seq, num int64) ([]*fan.UserInfo, int64) {
	table := "fan"
	if rtype == followerType {
		table = "follower"
	}
	query := fmt.Sprintf("SELECT r.id, r.tuid, u.headurl, u.nickname FROM %s r, users u WHERE r.tuid = u.uid AND r.uid = %d ", table, uid)
	if seq != 0 {
		query += fmt.Sprintf(" AND r.id < %d", seq)
	}
	query += fmt.Sprintf(" ORDER BY r.id DESC LIMIT %d", num)

	var infos []*fan.UserInfo
	var nextseq int64
	rows, err := db.Query(query)
	if err != nil {
		log.Printf("getRelations query failed:%v", err)
		return infos, nextseq
	}

	defer rows.Close()
	for rows.Next() {
		var info fan.UserInfo
		err := rows.Scan(&nextseq, &info.Uid, &info.Headurl, &info.Nickname)
		if err != nil {
			log.Printf("getRelations scan failed:%v", err)
			continue
		}
		infos = append(infos, &info)
	}
	return infos, nextseq
}
