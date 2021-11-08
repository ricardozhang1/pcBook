gen:
	protoc --go-grpc_out=.  --go_out=. --proto_path=proto proto/*.proto

clean:
	rm pd/*.go

server:
	go run cmd/server/main.go -port 8080

client:
	go run cmd/client/main.go -address 0.0.0.0:8080

test:
	go test -cover -race ./...


.PHONY: clean gen server test client