package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"rate-limiter-service/handlers"
	pb "rate-limiter-service/proto/ratelimiter"

	"rate-limiter-service/config"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {

	cfg := config.GetConfig()

	shutdownTracing := initTracing(
		"rate-limiter-service",
		cfg.JaegerHost+":"+cfg.JaegerGRPCPort,
	)
	defer shutdownTracing()

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}
	defer func(listener net.Listener) {
		err := listener.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(listener)

	// Bootstrap gRPC server.
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
	)

	// Bootstrap gRPC service server and respond to request.
	rateLimiterHandler, err := handlers.NewRateLimiterHandler()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	// Attach the RateLimit service to the server
	pb.RegisterRateLimitServiceServer(grpcServer, rateLimiterHandler)
	reflection.Register(grpcServer)
	// Serve gRPC server
	log.Println("Serving gRPC on 0.0.0.0:8080")
	go func() {
		log.Fatalln(grpcServer.Serve(listener))
	}()

	// Create a client connection to the gRPC server we just started
	// This is where the gRPC-Gateway proxies the requests
	conn, err := grpc.DialContext(
		context.Background(),
		"0.0.0.0:8080",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalln("Failed to dial server:", err)
	}

	gwmux := runtime.NewServeMux()
	// Register RateLimit client
	err = pb.RegisterRateLimitServiceHandler(context.Background(), gwmux, conn)
	if err != nil {
		log.Fatalln("Failed to register gateway:", err)
	}

	gwServer := &http.Server{
		Addr:    ":8090",
		Handler: gwmux,
	}

	log.Println("Serving gRPC-Gateway on http://0.0.0.0:8090")
	log.Fatalln(gwServer.ListenAndServe())
}
