usage: FORCE
	# See targets in Makefile (e.g. "buildlet.darwin-amd64")
	exit 1

FORCE:

buildlet.darwin-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.freebsd-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.linux-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.netbsd-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.openbsd-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.openbsd-386: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.plan9-386: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.windows-amd64: FORCE buildlet_windows.go
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.linux-arm: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.linux-arm64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.linux-ppc64le: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

buildlet.solaris-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public go-builder-data/$@

dev-buildlet.linux-arm: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public dev-go-builder-data/$@

dev-buildlet.linux-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public dev-go-builder-data/$@

dev-buildlet.windows-amd64: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public dev-go-builder-data/$@

dev-buildlet.plan9-386: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=$@ --file=go:golang.org/x/build/cmd/buildlet --public dev-go-builder-data/$@
