#!/bin/bash
mkdir -p /data/log
mkdir -p /data/server
mkdir -p /data/rpc
mkdir -p /data/etcd
mkdir -p /data/init
mkdir -p /usr/local/gocode

yum install unzip
yum install psmisc

cd /data/init
#golang
wget https://storage.googleapis.com/golang/go1.8.1.linux-amd64.tar.gz
tar -zxvf go1.8.1.linux-amd64.tar.gz
mv go /usr/local
export PATH=$PATH:/usr/local/go/bin
export GOPATH=/usr/local/gocode:/data/darren

#protoc
wget https://github.com/google/protobuf/releases/download/v3.3.0/protoc-3.3.0-linux-x86_64.zip
mkdir -p /usr/local/protobuf
mv protoc-3.3.0-linux-x86_64.zip /usr/local/protobuf
cd /usr/local/protobuf/
unzip protoc-3.3.0-linux-x86_64.zip
export PATH=$PATH:/usr/local/protobuf/bin

#go get source
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get github.com/facebookgo/grace/gracehttp
go get github.com/bitly/go-simplejson
go get golang.org/x/net/context
go get google.golang.org/grpc
go get github.com/nsqio/go-nsq
go get github.com/coreos/etcd/clientv3
go get gopkg.in/redis.v5
go get github.com/go-sql-driver/mysql
go get github.com/mercari/go-grpc-interceptor/panichandler
github.com/satori/go.uuid

#etcd
cd /data/etcd
wget https://github.com/coreos/etcd/releases/download/v3.2.0-rc.0/etcd-v3.2.0-rc.0-linux-amd64.tar.gz
tar -zxvf etcd-v3.2.0-rc.0-linux-amd64.tar.gz
mv etcd-v3.2.0-rc.0-linux-amd64 /usr/local/etcd
export PATH=$PATH:/usr/local/etcd/
go get github.com/mattn/goreman
nohup goreman start &

#nsq
cd /data/init
wget https://s3.amazonaws.com/bitly-downloads/nsq/nsq-1.0.0-compat.linux-amd64.go1.8.tar.gz
tar -zxvf nsq-1.0.0-compat.linux-amd64.go1.8.tar.gz
mv nsq-1.0.0-compat.linux-amd64.go1.8 /usr/local/nsq
nohup /usr/local/nsq/bin/nsqlookupd 1>/data/log/nsqlookupd.log 2>&1 &
nohup /usr/local/nsq/bin/nsqd 1>/data/log/nsqd.log 2>&1 &
