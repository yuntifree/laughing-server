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

func getFollowerUids(db *sql.DB, uid int64) map[int64]bool {
	m := make(map[int64]bool)
	rows, err := db.Query("SELECT tuid FROM follower WHERE uid = ? AND deleted = 0 ", uid)
	if err != nil {
		return m
	}
	defer rows.Close()
	for rows.Next() {
		var uid int64
		err = rows.Scan(&uid)
		if err != nil {
			continue
		}
		m[uid] = true
	}
	return m
}

func getRelations(db *sql.DB, uid, rtype, seq, num int64) ([]*fan.UserInfo, int64) {
	table := "fan"
	if rtype == followerType {
		table = "follower"
	}
	query := fmt.Sprintf("SELECT r.id, r.tuid, u.headurl, u.nickname FROM %s r, users u WHERE r.tuid = u.uid AND r.uid = %d AND r.deleted = 0 ", table, uid)
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
	var followers map[int64]bool
	if rtype == fanType {
		followers = getFollowerUids(db, uid)
		log.Printf("followers:%v", followers)
	}

	defer rows.Close()
	for rows.Next() {
		var info fan.UserInfo
		err := rows.Scan(&nextseq, &info.Uid, &info.Headurl, &info.Nickname)
		if err != nil {
			log.Printf("getRelations scan failed:%v", err)
			continue
		}
		if rtype == fanType {
			if len(followers) > 0 {
				if _, ok := followers[info.Uid]; ok {
					info.Hasfollow = 1
				} else {
					info.Hasfollow = 0
				}
			} else {
				info.Hasfollow = 0
			}
		}
		infos = append(infos, &info)
	}
	return infos, nextseq
}
