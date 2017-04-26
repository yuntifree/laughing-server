package util

import (
	"laughing-server/proto/fan"
	"laughing-server/proto/modify"
	"laughing-server/proto/share"
	"laughing-server/proto/user"
	"laughing-server/proto/verify"

	"google.golang.org/grpc"
)

const (
	ErrInvalidParam = 2
)

//GenServerName generate rpc server name
func GenServerName(rtype int64, callback string) string {
	switch rtype {
	case FanServerType:
		return FanServerName
	case VerifyServerType:
		return VerifyServerName
	case ShareServerType:
		return ShareServerName
	case UserServerType:
		return UserServerName
	case ModifyServerType:
		return ModifyServerName
	default:
		panic(AppError{ErrInvalidParam, "illegal server type", callback})
	}
}

//GenClient generate rpc client
func GenClient(rtype int64, conn *grpc.ClientConn, callback string) interface{} {
	var cli interface{}
	switch rtype {
	case FanServerType:
		cli = fan.NewFanClient(conn)
	case VerifyServerType:
		cli = verify.NewVerifyClient(conn)
	case ShareServerType:
		cli = share.NewShareClient(conn)
	case UserServerType:
		cli = user.NewUserClient(conn)
	case ModifyServerType:
		cli = modify.NewModifyClient(conn)
	default:
		panic(AppError{ErrInvalidParam, "illegal server type", callback})
	}
	return cli
}
