#!/bin/bash
RPCLIST="discover fan verify user modify limit config"
SPECLIST="share"

function gen_rpc()
{
    for srv in $RPCLIST; do
        echo $srv
        protoc --go_out=plugins=grpc:../proto/$srv/ ../proto/$srv/$srv.proto -I../.. -I../proto/$srv
     done
}

function gen_spec()
{
    for srv in $SPECLIST; do
        echo $srv
        protoc --go_out=plugins=grpc:../proto/$srv/ ../proto/$srv/$srv.proto -I../.. -I../proto/$srv
        sed -i "s/\"uid,omitempty\"/\"uid\"/g" ../proto/$srv/$srv.pb.go
     done
}

function gen_comm()
{
    protoc --go_out=../proto/common ../proto/common/common.proto -I../proto/common
}

gen_comm
gen_rpc
gen_spec
