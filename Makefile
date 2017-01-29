.PHONY: test
test: vendor
	go test ./util

vendor: glide.yaml
	go get -u github.com/Masterminds/glide
	glide install
