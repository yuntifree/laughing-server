package util

const (
	DiscoverServerName = "service:discover"
	DiscoverServerPort = ":60060"
	FanServerName      = "service:fan"
	FanServerPort      = ":60061"
	VerifyServerName   = "service:laughverify"
	VerifyServerPort   = ":60062"
	ShareServerName    = "service:share"
	ShareServerPort    = ":60063"
	UserServerName     = "service:user"
	UserServerPort     = ":60064"
	ModifyServerName   = "service:laughmodify"
	ModifyServerPort   = ":60065"
	LimitServerName    = "service:limit"
	LimitServerPort    = ":60066"
	ConfigServerName   = "service:laughconfig"
	ConfigServerPort   = ":60067"
	MaxIdleConns       = 3
	DebugHost          = "10.8.10.57"
	APIHosts           = "10.11.38.52"
	TimeFormat         = "2006-01-02 15:04:05"
	RpcConfPath        = "/data/laughing/rpc/rpc.conf"
	DiscoverServerType = 1
	FanServerType      = 2
	VerifyServerType   = 3
	ShareServerType    = 4
	UserServerType     = 5
	ModifyServerType   = 6
	LimitServerType    = 7
	ConfigServerType   = 8
)
