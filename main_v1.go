// main.go
package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	pb "myproject/example" // замените на актуальный путь к сгенерированному пакету
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Создаём HTTP-мультиплексор, который будет принимать REST-запросы
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	// Регистрируем хендлер для gRPC-сервиса, реализуемого на Node.js
	err := pb.RegisterExampleServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts)
	if err != nil {
		return err
	}

	log.Println("gRPC Gateway listening on :8080")
	return http.ListenAndServe(":8080", mux)
}

func main() {
	flag.Parse()
	if err := run(); err != nil {
		log.Fatalf("Failed to run gateway: %v", err)
	}
}
