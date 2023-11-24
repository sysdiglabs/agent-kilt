agent-kilt.zip: cmd/handler/handler
	cd cmd/handler/ && zip ../../agent-kilt.zip ./handler

cmd/handler/handler:
	cd cmd/handler && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build .

clean:
	rm agent-kilt.zip || true
	rm cmd/handler/handler || true

.PHONY: clean
