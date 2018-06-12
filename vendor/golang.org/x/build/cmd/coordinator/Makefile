usage: FORCE
	# See Makefile targets.
	exit 1

FORCE:

# After "make upload", either reboot the machine, or ssh to it and:
#   sudo systemctl restart gobuild.service
# And watch its logs with:
#   sudo journalctl -f -u gobuild.service
upload: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=linux-amd64 --file=go:golang.org/x/build/cmd/coordinator --public go-builder-data/coordinator

dev-upload: FORCE
	go install golang.org/x/build/cmd/upload
	upload --verbose --osarch=linux-amd64 --file=go:golang.org/x/build/cmd/coordinator --public dev-go-builder-data/coordinator
