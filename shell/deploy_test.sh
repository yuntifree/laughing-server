RPCLIST="discover modify verify user fan limit share config"
HTTPLIST="appserver ossserver"
CRONLIST="fetchHead regFollow fetchVideoInfo loadAPIStat loadRPCStat"
for srv in $HTTPLIST; do
    go build ../access/$srv
    ./install.sh 1 $srv
done

for srv in $RPCLIST; do
    go build ../rpc/$srv
    ./install.sh 2 $srv
done

for srv in $CRONLIST; do
    go build ../tools/$srv.go
    ./install.sh 3 $srv
done
