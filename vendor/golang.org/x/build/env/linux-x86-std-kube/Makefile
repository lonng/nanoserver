# Copyright 2015 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
IMAGE_NAME=$(shell basename $(CURDIR))
STAGING_REPO=gcr.io/go-dashboard-dev
PROD_REPO=gcr.io/symbolic-datum-552

usage:
	echo "Use staging, prod, or dev targets. For dev, specify your Docker repository with the REPO=foo argument." ; exit 1

staging: Dockerfile
	docker build -t $(STAGING_REPO)/$(IMAGE_NAME):latest .
	gcloud docker push $(STAGING_REPO)/$(IMAGE_NAME):latest

prod: Dockerfile
	docker build -t $(PROD_REPO)/$(IMAGE_NAME):latest .
	gcloud docker push $(PROD_REPO)/$(IMAGE_NAME):latest

# You must provide a REPO=your-repo-name arg when you make
# this targarget. REPO is the name of the Docker repository
# that will be prefixed to the name of the image being built.
dev: Dockerfile
	docker build -t $(REPO)/$(IMAGE_NAME):latest .
	gcloud docker push $(REPO)/$(IMAGE_NAME):latest
