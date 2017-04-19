package util

import "database/sql"

const (
	//MaxListSize for page
	MaxListSize = 20
	masterRds   = "rm-wz9sb2613092ki9xn.mysql.rds.aliyuncs.com"
	readRds     = "rm-wz9sb2613092ki9xn.mysql.rds.aliyuncs.com"
)

func genDsn(readonly bool) string {
	host := masterRds
	if readonly {
		host = readRds
	}
	return "access:^yunti9df3b01c$@tcp(" + host + ":3306)/yunxing?charset=utf8"
}

//InitDB connect to rds
func InitDB(readonly bool) (*sql.DB, error) {
	dsn := genDsn(readonly)
	return sql.Open("mysql", dsn)
}
