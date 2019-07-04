#!/bin/sh

export GOPROXY=https://goproxy.io
export GOOS=linux
export GOARCH=amd64

echo "============================="
echo "==== building"
echo "============================="
go build -o mahjong

if [[ $? -ne 0 ]]
then
    echo "build failed"
    exit -1
fi

echo "============================="
echo "==== packaging"
echo "============================="
tar -czf mahjong.tar.gz mahjong configs

rm -rf dist
mkdir -p dist
mv mahjong.tar.gz dist/

echo "============================="
echo "==== clean"
echo "============================="
rm -rf "mahjong"
