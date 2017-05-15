package util

import "database/sql"

const (
	//MaxListSize for page
	MaxListSize = 20
	masterRds   = "10.11.56.116"
	readRds     = "10.11.56.116"
	access      = "root:^laughingFxT@#$"
)

func genDsn(readonly bool) string {
	host := masterRds
	if readonly {
		host = readRds
	}
	return access + "@tcp(" + host + ":3306)/laughing?charset=utf8"
}

func genMonitorDsn() string {
	return access + "@tcp(" + masterRds + ":3306)/monitor?charset=utf8"
}

//InitDB connect to rds
func InitDB(readonly bool) (*sql.DB, error) {
	dsn := genDsn(readonly)
	return sql.Open("mysql", dsn)
}

//InitDBParam init mysql connection with params
func InitDBParam(access, host string) (*sql.DB, error) {
	dsn := access + "@tcp(" + host + ":3306)/laughing?charset=utf8"
	return sql.Open("mysql", dsn)
}

//InitMonitorDB connect to rds
func InitMonitorDB() (*sql.DB, error) {
	dsn := genMonitorDsn()
	return sql.Open("mysql", dsn)
}
