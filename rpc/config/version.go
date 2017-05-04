package main

import (
	"database/sql"
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
