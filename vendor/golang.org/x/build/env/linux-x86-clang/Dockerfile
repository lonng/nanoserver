# Copyright 2014 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Linux builder VM with clang instead of gccc.
# Docker tag gobuilders/linux-x86-clang

FROM debian:wheezy
MAINTAINER golang-dev <golang-dev@googlegroups.com>

ENV DEBIAN_FRONTEND noninteractive

ADD /sources/clang-deps.list /etc/apt/sources.list.d/

RUN apt-get update
RUN apt-get install -y --no-install-recommends ca-certificates
# Optionally used by some net/http tests:
RUN apt-get install -y --no-install-recommends strace
# For building Go's bootstrap 'dist' prog
RUN apt-get install -y --no-install-recommends curl
RUN curl http://llvm.org/apt/llvm-snapshot.gpg.key | apt-key add -

RUN apt-get update
RUN apt-get install -y --no-install-recommends clang-3.5
# TODO(cmang): move these into a 386 image that derives from this one.
RUN apt-get install -y --no-install-recommends libc6-dev-i386 gcc-multilib
# Remove gcc binary so it doesn't interfere with clang
RUN rm -f /usr/bin/gcc

# Required for networking in a VM on GCE:
RUN apt-get install -y --no-install-recommends net-tools ifupdown isc-dhcp-client dhcp3-client
# And misc basic tools:
RUN apt-get install -y --no-install-recommends procps lsof psmisc

RUN apt-get clean

RUN mkdir -p /go1.4-amd64 && (curl --silent https://storage.googleapis.com/golang/go1.4.linux-amd64.tar.gz | tar -C /go1.4-amd64 -zxv)
RUN mv /go1.4-amd64/go /go1.4 && rmdir /go1.4-amd64

RUN curl -o /usr/local/bin/stage0 https://storage.googleapis.com/go-builder-data/stage0.linux-amd64.92ebb4ec
RUN chmod +x /usr/local/bin/stage0

ADD scripts/rc.local /etc/init.d/rc.local
RUN chmod +x /etc/init.d/rc.local

RUN rm -rf /var/lib/apt/lists /usr/share/doc
RUN rm -rf /go1.4/pkg/linux_amd64_race /go1.4/api /go1.4/blog /go1.4/doc /go1.4/misc /go1.4/test
RUN find /go1.4 -type d -name testdata | xargs rm -rf
RUN rm -rf /go1.4/pkg/linux_amd64_race /go1.4/test /go1.4/api
RUN (cd /usr/share/locale/ && ls -1 | grep -v en | xargs rm -rf)
RUN rm -rf /var/cache/debconf/*
RUN rm -rf /usr/share/man

RUN (echo "auto lo"; echo "iface lo inet loopback"; echo "auto eth0"; echo "iface eth0 inet dhcp") > /etc/network/interfaces
