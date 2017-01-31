.PHONY: test
test: vendor .test-image-flag
	docker run -it \
		-v $(PWD):/go/src/github.com/memsql/online-upgrade \
		-w /go/src/github.com/memsql/online-upgrade \
		memsql-online-upgrade \
		go test $(shell glide nv)

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
		memsql-online-upgrade
		/bin/bash
