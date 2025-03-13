go install .\main_v2.go
go mod tidy

protoc -I. -I./googleapis --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative hello/hello.proto

protoc -I. -I./googleapis --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --grpc-gateway_out=. --grpc-gateway_opt=paths=source_relative auth/auth.proto

go run .\main_v2.go

npm i
node .\hello\
node .\auth