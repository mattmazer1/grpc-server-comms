
compile:
	protoc --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative  proto/service.proto

startServer:
	go run ./server/main.go

startClient:
	go run  ./client/client.go

