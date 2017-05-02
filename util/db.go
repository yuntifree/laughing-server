package util

import "database/sql"

const (
	//MaxListSize for page
	MaxListSize = 20
	masterRds   = "10.11.56.116"
	readRds     = "10.11.56.116"
)

func genDsn(readonly bool) string {
	host := masterRds
	if readonly {
		host = readRds
	}
	return "root:^laughingFxT@#$@tcp(" + host + ":3306)/laughing?charset=utf8"
}

//InitDB connect to rds
func InitDB(readonly bool) (*sql.DB, error) {
	dsn := genDsn(readonly)
	return sql.Open("mysql", dsn)
}
