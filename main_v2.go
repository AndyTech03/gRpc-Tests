package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"google.golang.org/grpc"
	//"google.golang.org/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/metadata"

	pbAuth "myproject/auth"
	pbHello "myproject/hello"
)

func authMiddleware(ctx context.Context, req *http.Request) metadata.MD {
	if req.URL.Path == "/v2/login" /*|| req.URL.Path == "/v2/validate"*/ {
		return nil
	}

	accessToken := req.Header.Get("Authorization")
	if accessToken == "" {
		return metadata.Pairs("err", "Missing token", "errCode", "403")
	}

	conn, err := grpc.Dial("localhost:50052", grpc.WithInsecure())
	log.Println(err)
	if err != nil {
		log.Println(err)
		return metadata.Pairs("err", "Auth failed. Try again.", "errCode", "500")
	}
	defer conn.Close()

	authClient := pbAuth.NewAuthServiceClient(conn)
	resp, err := authClient.ValidateToken(ctx, &pbAuth.ValidateRequest{AccessToken: accessToken})
	if err != nil {
		log.Println(err)
		return metadata.Pairs("err", "Auth failed. Try again.", "errCode", "500")
	}
	if !resp.Valid {
		return metadata.Pairs("err", "Invalid token", "errCode", "403")
	}

	md := metadata.Pairs("username", resp.Username)
	// newCtx := metadata.NewOutgoingContext(ctx, md)
	return md
}

func customHTTPErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
    st, ok := status.FromError(err)
    if !ok {
        // Если не удалось получить статус ошибки, считаем это внутренней ошибкой
        w.WriteHeader(http.StatusInternalServerError)
        w.Write([]byte(err.Error()))
        return
    }
    // По умолчанию, runtime.HTTPStatusFromCode сопоставляет коды правильно:
    // PermissionDenied (7) -> 403, Unauthenticated (16) -> 401 и т.д.
    httpStatus := runtime.HTTPStatusFromCode(convertHTTPToGRPC(st.Code()))
    w.WriteHeader(httpStatus)
    jsonErr, _ := marshaler.Marshal(map[string]interface{}{
        "code":    st.Code(),
        "message": st.Message(),
    })
    w.Write(jsonErr)
}

func convertHTTPToGRPC(httpCode codes.Code) codes.Code {
	switch httpCode {
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.Aborted
	case http.StatusInternalServerError:
		return codes.Internal
	default:
		return codes.Unknown
	}
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux(
		runtime.WithMetadata(authMiddleware),
		runtime.WithErrorHandler(customHTTPErrorHandler),
	)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := pbAuth.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, "localhost:50052", opts); err != nil {
		return err
	}

	if err := pbHello.RegisterHelloServiceHandlerFromEndpoint(ctx, mux, "localhost:50051", opts); err != nil {
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
