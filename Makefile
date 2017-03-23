# If no test provided, run all.
ifndef $(test)
	test=$(shell glide nv)
endif

.PHONY: test test-image-shell vendor
# `make test` runs all tests $(shell glide nv) in the project.
# `-p 1` is required when running multiple tests that build a
# cluster. Without it, parallelization can occur causing
# clusters to be built simultaneously and fail.
# `make test` will build an HA cluster and will also run 
# docker-clean and docker network first.
# 
# Alternatively, you can provide a single package to test.
# e.g. make test test='./util/ops_test.go'
test: docker-clean docker-network vendor build-memsql 
	docker run -di --network=memsql --name=online-upgrade-child \
		-v $(PWD):/go/src/github.com/memsql/online-upgrade \
		-w /go/src/github.com/memsql/online-upgrade \
		memsql-online-upgrade
	docker run -t -i --network=memsql --name=online-upgrade-master \
		-v $(PWD):/go/src/github.com/memsql/online-upgrade \
		-w /go/src/github.com/memsql/online-upgrade \
		memsql-online-upgrade \
		go test -p 1 -v -timeout 30m $(test)

# Build the MemSQL Online Upgrade container
build-memsql: Dockerfile
	docker build -t memsql-online-upgrade .

# Install go application dependencies
vendor: glide.yaml
	go get -u github.com/Masterminds/glide
	glide install

# Setup custom docker network
# This is required to use --name and --hostname with Docker run
docker-network:
	docker network create -d bridge memsql

# Test Image Shell spins up both master and child continers
# You can ssh between the container as root using the password 
# specified in the Dockerfile. (ssh root@online-upgrade-child)
test-image-shell: docker-clean docker-network vendor
	docker run -di --network=memsql --name=online-upgrade-child \
		-v $(PWD):/go/src/github.com/memsql/online-upgrade \
		-w /go/src/github.com/memsql/online-upgrade \
		memsql-online-upgrade
	docker run -t -i --network=memsql --name=online-upgrade-master \
		-v $(PWD):/go/src/github.com/memsql/online-upgrade \
		-w /go/src/github.com/memsql/online-upgrade \
		memsql-online-upgrade \
		/bin/bash
		
# Clean up containers from previous test run and remove the custom network
docker-clean:
	-docker rm -f  online-upgrade-master online-upgrade-child
	-docker network rm memsql
