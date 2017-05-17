#!/bin/bash
# Need some system function.
. /etc/init.d/functions

OPDIR=/data/laughing/oss
HTMLDIR=/data/laughing/html
IPLIST="107.150.97.203"

function install_html()
{
    for ip in $IPLIST; do
        scp -r $1 root@$ip:/tmp
        ssh root@$ip "rm -rf $HTMLDIR/$1_bak"
        ssh root@$ip "mv -f $HTMLDIR/$1 $HTMLDIR/$1_bak"
        ssh root@$ip "mv -f /tmp/$1 $HTMLDIR"
    done
}

function install_oss()
{
    for ip in $IPLIST; do
        scp -r $1 root@$ip:/tmp
        ssh root@$ip "rm -rf $OPDIR/$1_bak"
        ssh root@$ip "mv -f $OPDIR/$1 $OPDIR/$1_bak"
        ssh root@$ip "mv -f /tmp/$1 $OPDIR"
    done
}

arr=$*
args=${arr[@]:2}

for arg in $args
do
    if [ $1 -eq 1 ]; then
        install_html $arg
    elif [ $1 -eq 2 ]; then
        install_oss $arg
    else
        echo "illegal type"
    fi
done
