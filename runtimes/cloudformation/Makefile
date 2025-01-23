GOOS?=linux
GOARCH?= amd64

cmd/handler/handler:
	cd cmd/handler && CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build .

clean:
	rm cmd/handler/handler || true

test:
	go test -v ./...

.PHONY: clean test
