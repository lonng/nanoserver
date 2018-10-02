#!/bin/sh

BASE=$(dirname $0)/..

$BASE/build/build.sh

if [ $? -ne 0 ]
then
    echo "build failed"
    exit -1
fi

REMOTE=YOUR_REMOTE_SERVER_IP

echo "============================="
echo "==== deploy to remote server"
echo "============================="
scp dist/mahjong.tar.gz root@$REMOTE:/opt/mahjong

ssh root@$REMOTE <<EOF
cd /opt/mahjong
tar -xzf mahjong.tar.gz
chmod +x mahjong
ls -al
supervisorctl restart mahjong
supervisorctl status
EOF

echo "done"
