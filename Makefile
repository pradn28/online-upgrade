.PHONY: test
# "test" Runs each test $(shell glide nv) in Parallel (-p 1)
# -p 1 is required when running multiple tests that build a
# cluster. Without it, parallelization can occur causing
# clusters to be built simultaneously and fail
test: vendor .test-image-flag
	docker run -it \
		-v $(PWD):/go/src/github.com/memsql/online-upgrade \
		-w /go/src/github.com/memsql/online-upgrade \
		memsql-online-upgrade \
		go test -p 1 $(shell glide nv)
	# Clean up containers after we are all done
	docker ps -aq | xargs sudo docker rm \

# Provide a single package to test (make test-one test='./util/...')
test-one: vendor .test-image-flag
	docker run -it \
		-v $(PWD):/go/src/github.com/memsql/online-upgrade \
		-w /go/src/github.com/memsql/online-upgrade \
		memsql-online-upgrade \
		go test $(test)

.test-image-flag: Dockerfile
	docker build -t memsql-online-upgrade .
	touch .test-image-flag

vendor: glide.yaml
	go get -u github.com/Masterminds/glide
	glide install

.PHONY: test-image-shell
test-image-shell:
	docker run -it \
		-v $(PWD):/go/src/github.com/memsql/online-upgrade \
		-w /go/src/github.com/memsql/online-upgrade \
		memsql-online-upgrade \
		/bin/bash
