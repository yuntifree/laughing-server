#!/bin/bash

HTTPSRV=(appserver)
RPCSRV=(discover config modify verify user fan limit share)
CRONLIST=(regFollow loadAPIStat loadRPCStat fetchHead fetchVideoInfo)
LOG=/var/log/srv.log
ERR=/var/log/srv.err


HTTP_DIR=/data/laughing/server
RPC_DIR=/data/laughing/rpc
CRON_DIR=/data/laughing/cron

function log()
{
    echo "[$(date "+%F %T"),000] [agent]$1" >> $LOG
    echo "$1"
}

function err()
{
    echo "[$(date "+%F %T"),000] [agent]$1" >> $ERR
    echo "$1"
}

function pullhttp()
{
    nohup $HTTP_DIR/$1 1>>$HTTP_DIR/$1.log 2>&1 &
}

function pullrpc()
{
    nohup $RPC_DIR/$1 1>>$RPC_DIR/$1.log 2>&1 &
}

function pullcron()
{
    nohup $CRON_DIR/$1 1>>$CRON_DIR/$1.log 2>&1 &
}

function check_http()
{
    for srv in ${HTTPSRV[@]}; do
        sname=$HTTP_DIR/$srv
        if [ -z "$(ps -ef |grep $sname| grep -v grep|grep -v $sname.log)" ]; then
            err "Server $sname not running, restart."
            pullhttp $srv
        fi
    done
}

function check_rpc()
{
    for srv in ${RPCSRV[@]}; do
        sname=$RPC_DIR/$srv
        if [ -z "$(ps -ef |grep $sname| grep -v grep|grep -v $sname.log)" ]; then
            err "Server $sname not running, restart."
            pullrpc $srv
        fi
    done
}

function check_cron()
{
    for srv in ${CRONLIST[@]}; do
        sname=$CRON_DIR/$srv
        if [ -z "$(ps -ef |grep $sname| grep -v grep|grep -v $sname.log)" ]; then
            err "Server $sname not running, restart."
            pullcron $srv
        fi
    done
}

check_http
check_rpc
check_cron
