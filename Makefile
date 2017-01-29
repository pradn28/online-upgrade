.PHONY: test
test: vendor
	go test $(shell glide nv)

vendor: glide.yaml
	go get -u github.com/Masterminds/glide
	glide install
