

.PHONY: install-dependencies
install-dependencies:
	go get github.com/gogo/protobuf/protoc-gen-gofast

.PHONY: proto
proto:
	@protoc --gofast_out=plugins=grpc:. ./event/event.proto
