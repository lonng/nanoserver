# Copyright 2014 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Linux builder VM running Debian sid (i.e. Debian tip)
# Docker tag gobuilders/linux-x86-sid

FROM debian:sid
MAINTAINER golang-dev <golang-dev@googlegroups.com>

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update
RUN apt-get install -y --no-install-recommends ca-certificates
# Optionally used by some net/http tests:
RUN apt-get install -y --no-install-recommends strace
# Optionally used by runtime tests for gdb:
RUN apt-get install -y --no-install-recommends gdb
# For building Go's bootstrap 'dist' prog
RUN apt-get install -y --no-install-recommends gcc libc6-dev
# For 32-bit builds:
# TODO(bradfitz): move these into a 386 image that derives from this one.
RUN apt-get install -y --no-install-recommends libc6-dev-i386 gcc-multilib
# For interacting with the Go source & subrepos:
RUN apt-get install -y --no-install-recommends git-core
# For installing Go 1.4:
RUN apt-get install -y --no-install-recommends curl

# Required for networking in a VM on GCE:
RUN apt-get install -y --no-install-recommends net-tools ifupdown isc-dhcp-client
# And misc basic tools:
RUN apt-get install -y --no-install-recommends procps lsof psmisc

RUN apt-get clean
RUN rm -fr /var/lib/apt/lists

RUN mkdir -p /go1.4-amd64 && (curl --silent https://storage.googleapis.com/golang/go1.4.linux-amd64.tar.gz | tar -C /go1.4-amd64 -zxv)
RUN mv /go1.4-amd64/go /go1.4 && rmdir /go1.4-amd64

RUN curl -o /usr/local/bin/stage0 https://storage.googleapis.com/go-builder-data/stage0.linux-amd64.92ebb4ec
RUN chmod +x /usr/local/bin/stage0

ADD buildlet.service /etc/systemd/system/buildlet.service
RUN systemctl enable /etc/systemd/system/buildlet.service

RUN rm -rf /var/lib/apt/lists /usr/share/doc
RUN rm -rf /go1.4/pkg/linux_amd64_race /go1.4/api /go1.4/blog /go1.4/doc /go1.4/misc /go1.4/test
RUN find /go1.4 -type d -name testdata | xargs rm -rf
RUN rm -rf /go1.4/pkg/linux_amd64_race /go1.4/test /go1.4/api
RUN (cd /usr/share/locale/ && ls -1 | grep -v en | xargs rm -rf)
RUN rm -rf /var/cache/debconf/*
RUN rm -rf /usr/share/man

RUN rm -f /lib/systemd/system/multi-user.target.wants/systemd-logind.service
RUN (echo "[Journal]"; echo "ForwardToConsole=yes") > /etc/systemd/journald.conf

RUN (echo "auto lo"; echo "iface lo inet loopback"; echo "auto eth0"; echo "iface eth0 inet dhcp") > /etc/network/interfaces
