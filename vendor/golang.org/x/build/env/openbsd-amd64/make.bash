#!/bin/bash
# Copyright 2014 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# This script requires expect, growisofs and qemu.

set -e
set -u

readonly ARCH="${ARCH:-amd64}"
readonly MIRROR="${MIRROR:-ftp.usa.openbsd.org}"

if [[ "${ARCH}" != "amd64" && "${ARCH}" != "i386" ]]; then
  echo "ARCH must be amd64 or i386"
  exit 1
fi

readonly ISO="install58-${ARCH}.iso"
readonly ISO_PATCHED="install58-${ARCH}-patched.iso"

if [[ ! -f "${ISO}" ]]; then
  curl -o "${ISO}" "http://${MIRROR}/pub/OpenBSD/5.8/${ARCH}/install58.iso"
fi

function cleanup() {
	rm -f "${ISO_PATCHED}"
	rm -f auto_install.conf
	rm -f boot.conf
	rm -f disk.raw
	rm -f disklabel.template
	rm -f etc/rc.local
	rm -f install.site
	rm -f random.seed
	rm -f site58.tgz
	rmdir etc
}

trap cleanup EXIT INT

# XXX: Download and save bash, curl, and their dependencies too?
# Currently we download them from the network during the install process.

# Create custom site58.tgz set.
mkdir -p etc
cat >install.site <<EOF
#!/bin/sh
env PKG_PATH=http://${MIRROR}/pub/OpenBSD/5.8/packages/${ARCH} \
  pkg_add -iv bash curl git

echo 'set tty com0' > boot.conf
EOF

cat >etc/rc.local <<EOF
(
  set -x
  echo "starting buildlet script"
  netstat -rn
  cat /etc/resolv.conf
  dig metadata.google.internal
  (
    set -e
    export PATH="\$PATH:/usr/local/bin"
    /usr/local/bin/curl -o /buildlet \$(/usr/local/bin/curl -H "Metadata-Flavor: Google" http://metadata.google.internal/computeMetadata/v1/instance/attributes/buildlet-binary-url)
    chmod +x /buildlet
    exec /buildlet
  )
  echo "giving up"
  sleep 10
  halt -p
)
EOF
chmod +x install.site
tar -zcvf site58.tgz install.site etc/rc.local

# Autoinstall script.
cat >auto_install.conf <<EOF
System hostname = buildlet
Which network interface = vio0
IPv4 address for vio0 = dhcp
IPv6 address for vio0 = none
DNS nameservers = 8.8.8.8
Password for root account = root
Do you expect to run the X Window System = no
Change the default console to com0 = yes
Which speed should com0 use = 115200
Setup a user = gopher
Full name for user gopher = Gopher Gopherson
Password for user gopher = gopher
Allow root ssh login = no
What timezone = US/Pacific
Which disk = sd0
Use (W)hole disk or (E)dit the MBR = whole
Use (A)uto layout, (E)dit auto layout, or create (C)ustom layout = auto
URL to autopartitioning template for disklabel = file://disklabel.template
Set name(s) = +* -x* -game* -man* done
Directory does not contain SHA256.sig. Continue without verification = yes
EOF

# Disklabel template.
cat >disklabel.template <<EOF
/	5G-*	95%
swap	1G
EOF

# Hack install CD a bit.
echo 'set tty com0' > boot.conf
dd if=/dev/urandom of=random.seed bs=4096 count=1
cp "${ISO}" "${ISO_PATCHED}"
growisofs -M "${ISO_PATCHED}" -l -R -graft-points \
  /5.8/${ARCH}/site58.tgz=site58.tgz \
  /auto_install.conf=auto_install.conf \
  /disklabel.template=disklabel.template \
  /etc/boot.conf=boot.conf \
  /etc/random.seed=random.seed

# Initialize disk image.
rm -f disk.raw
qemu-img create -f raw disk.raw 10G

# Run the installer to create the disk image.
expect <<EOF
spawn qemu-system-x86_64 -nographic -smp 2 -drive if=virtio,file=disk.raw \
  -cdrom "${ISO_PATCHED}" -net nic,model=virtio -net user -boot once=d

expect "boot>"
send "\n"

# Need to wait for the kernel to boot.
expect -timeout 600 "\(I\)nstall, \(U\)pgrade, \(A\)utoinstall or \(S\)hell\?"
send "s\n"

expect "# "
send "mount /dev/cd0c /mnt\n"
send "cp /mnt/auto_install.conf /mnt/disklabel.template /\n"
send "umount /mnt\n"
send "exit\n"

expect -timeout 600 "CONGRATULATIONS!"

expect -timeout 600 eof
EOF

# Create Compute Engine disk image.
echo "Archiving disk.raw... (this may take a while)"
tar -Szcf "openbsd-${ARCH}-gce.tar.gz" disk.raw

echo "Done. GCE image is openbsd-${ARCH}-gce.tar.gz."
