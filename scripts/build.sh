#!/bin/sh

#export GOROOT=/usr/local/Cellar/go/1.8rc
export GOOS=linux
export GOARCH=amd64
export BASEDIR=$(pwd)

echo "============================="
echo "==== building"
echo "============================="
cd "$BASEDIR/cmd/mahjong"
go build

if [ $? -ne 0 ]
then
    echo "build failed"
    exit -1
fi

echo "============================="
echo "==== packaging"
echo "============================="
cd "$BASEDIR/cmd/mahjong"
tar -czf mahjong.tar.gz mahjong configs web/static

cd $BASEDIR
rm -rf dist
mkdir -p dist
mv $BASEDIR/cmd/mahjong/mahjong.tar.gz dist/

echo "============================="
echo "==== clean"
echo "============================="
rm -rf "$BASEDIR/cmd/mahjong/mahjong"
