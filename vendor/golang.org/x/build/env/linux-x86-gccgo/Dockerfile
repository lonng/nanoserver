# Copyright 2014 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Linux builder VM for gccgo
# Docker tag gobuilders/linux-x86-gccgo

FROM debian:wheezy
MAINTAINER golang-dev <golang-dev@googlegroups.com>

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update
RUN apt-get install -y --no-install-recommends ca-certificates
# Optionally used by some net/http tests:
RUN apt-get install -y --no-install-recommends strace
# Optionally used by runtime tests for gdb:
RUN apt-get install -y --no-install-recommends gdb
# For using numeric libraries within GCC.
RUN apt-get install -y --no-install-recommends libgmp10-dev libmpc-dev libmpfr-dev
# For building gcc from source.
RUN apt-get install -y --no-install-recommends make g++ flex bison binutils-dev texinfo
# Same as above, but for 32-bit builds as well.
RUN apt-get install -y --no-install-recommends libc6-dev-i386 g++-multilib
# For running the extended gccgo testsuite
RUN apt-get install -y --no-install-recommends dejagnu
# For interacting with the gccgo source and git mirror:
RUN apt-get install -y --no-install-recommends git-core

# Required for networking in a VM on GCE:
RUN apt-get install -y --no-install-recommends curl net-tools ifupdown isc-dhcp-client dhcp3-client
# And misc basic tools:
RUN apt-get install -y --no-install-recommends procps lsof psmisc

RUN apt-get clean

RUN curl -o /usr/local/bin/stage0 https://storage.googleapis.com/go-builder-data/stage0.linux-amd64.92ebb4ec
RUN chmod +x /usr/local/bin/stage0

ADD scripts/rc.local /etc/init.d/rc.local
RUN chmod +x /etc/init.d/rc.local

RUN rm -rf /var/lib/apt/lists /usr/share/doc 
RUN (cd /usr/share/locale/ && ls -1 | grep -v en | xargs rm -rf)
RUN rm -rf /var/cache/debconf/*
RUN rm -rf /usr/share/man

RUN (echo "auto lo"; echo "iface lo inet loopback"; echo "auto eth0"; echo "iface eth0 inet dhcp") > /etc/network/interfaces