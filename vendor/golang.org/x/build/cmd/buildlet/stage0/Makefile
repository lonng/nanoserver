usage: FORCE
	# See targets in Makefile
	exit 1

FORCE:

buildlet-stage0.windows-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet/stage0 --public --cacheable=false go-builder-data/$@

buildlet-stage0.linux-arm-scaleway: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet/stage0 --public --cacheable=false go-builder-data/$@

buildlet-stage0.linux-arm64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet/stage0 --public --cacheable=false go-builder-data/$@

buildlet-stage0.linux-ppc64le: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet/stage0 --public --cacheable=false go-builder-data/$@

buildlet-stage0.solaris-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet/stage0 --public --cacheable=false go-builder-data/$@

# No staging vs prod targets here, because this file gets baked into
# the Docker layers on gcr.io (which do have staging vs prod names).
# That is, this file on GCS is only used at "docker build"-time, not
# during day-to-day builds.
buildlet-stage0.linux-amd64-kube: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet/stage0 --public --cacheable=false go-builder-data/$@

buildlet-stage0.linux-s390x: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet/stage0 --public --cacheable=false go-builder-data/$@
