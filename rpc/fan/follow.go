package main

import (
	"database/sql"
	"log"
)

const (
	FollowType   = 0
	UnfollowType = 1
)

func doFollow(db *sql.DB, uid, tuid int64) bool {
	res, err := db.Exec("INSERT INTO follower(uid, tuid, ctime) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE deleted = 0", uid, tuid)
	if err != nil {
		log.Printf("doFollow query failed:%v", err)
		return false
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		log.Printf("doFollow get rows affected failed:%v", err)
		return false
	}
	if cnt == 0 {
		log.Printf("%d has follow %d", uid, tuid)
		return true
	}
	_, err = db.Exec("INSERT INTO fan(uid, tuid, ctime) VALUES (?, ?, NOW()) ON DUPLICATE KEY UPDATE deleted = 0", tuid, uid)
	if err != nil {
		log.Printf("record fan failed:%d %d %v", tuid, uid, err)
		return false
	}
	_, err = db.Exec("UPDATE users SET follow_cnt = follow_cnt + 1 WHERE uid = ?", uid)
	if err != nil {
		log.Printf("update users follow_cnt failed:%d %v", uid, err)
		return false
	}
	_, err = db.Exec("UPDATE users SET fan_cnt = fan_cnt + 1 WHERE uid = ?", tuid)
	if err != nil {
		log.Printf("update users fan_cnt failed:%d %v", uid, err)
		return false
	}
	return true
}

func doUnfollow(db *sql.DB, uid, tuid int64) bool {
	res, err := db.Exec("UPDATE follower SET deleted = 1 WHERE uid = ? AND tuid = ?", uid, tuid)
	if err != nil {
		log.Printf("doUnfollow query failed:%d %d %v", uid, tuid, err)
		return false
	}
	cnt, err := res.RowsAffected()
	if err != nil {
		log.Printf("doFollow get rows affected failed:%v", err)
		return false
	}
	if cnt == 0 {
		log.Printf("%d has unfollow %d", uid, tuid)
		return true
	}
	_, err = db.Exec("UPDATE fan SET deleted = 1 WHERE uid = ? AND tuid = ?", tuid, uid)
	if err != nil {
		log.Printf("doUnfollow query failed:%d %d %v", uid, tuid, err)
		return false
	}
	_, err = db.Exec("UPDATE users SET follow_cnt = follow_cnt - 1 WHERE uid = ?", uid)
	if err != nil {
		log.Printf("update users follow_cnt failed:%d %v", uid, err)
		return false
	}
	_, err = db.Exec("UPDATE users SET fan_cnt = fan_cnt - 1 WHERE uid = ?", tuid)
	if err != nil {
		log.Printf("update users fan_cnt failed:%d %v", uid, err)
		return false
	}
	return true
}

func follow(db *sql.DB, otype, uid, tuid int64) bool {
	switch otype {
	case FollowType:
		return doFollow(db, uid, tuid)
	case UnfollowType:
		return doUnfollow(db, uid, tuid)
	}
	return false
}
