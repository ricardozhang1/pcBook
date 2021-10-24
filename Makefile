gen:
	protoc --go-grpc_out=:pd  --go_out=:pd --proto_path=proto proto/*.proto

clean:
	rm pd/*.go

server:
	go run cmd/server/main.go -port 8089

client:
	go run cmd/client/main.go -address 0.0.0.0:8089

test:
	go test -cover -race ./...


.PHONY: clean gen server test client