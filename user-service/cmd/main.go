package main

import (
	"context"
	"log"
	"net"
	"user-service/config"
	"user-service/internal/kafka"
	"user-service/internal/redis"
	"user-service/internal/service"
	"user-service/internal/utils"

	db "user-service/internal/storage"

	pb "user-service/protos/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

func main() {
	// Loading configuration (.env)
	config.LoadConfig()

	// Connecting to PostgreSQL
	dbConn, err := db.ConnectToDB(config.AppConfig)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	// Launching the Storage Layer
	dbPostgres := db.NewPostgresStorage(dbConn)
	redis := redis.NewRedisClient(config.AppConfig)
	kafka, err := kafka.NewProducer([]string{config.AppConfig.Kafka.Host + ":" + config.AppConfig.Kafka.Port})
	if err != nil {
		log.Fatalf("Kafka producer yaratishda xatolik: %v", err)
	}
	// Opening a TCP listener for the gRPC server
	lis, err := net.Listen("tcp", config.AppConfig.Http.Host+":"+config.AppConfig.Http.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// Creating a gRPC server and adding an interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcInterceptor),
	)
	// gRPC reflection (needed for tools like grpcurl)
	reflection.Register(grpcServer)

	// Registering a UserService server
	pb.RegisterUserServiceServer(grpcServer, service.NewUserService(dbPostgres, redis, kafka))

	// Starting the server
	log.Printf("User service running on %s:%s", config.AppConfig.Http.Host, config.AppConfig.Http.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// gRPC Interceptor: to log all RPC calls
func grpcInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req) // Token bo‘lmasa ham davom etsin (Register/Login uchun)
	}

	// Authorization sarlavhasidan tokenni olish
	authHeaders := md["authorization"]
	if len(authHeaders) > 0 {
		tokenStr := authHeaders[0]
		if len(tokenStr) > 7 && tokenStr[:7] == "Bearer " {
			tokenStr = tokenStr[7:] // Tokenni olib tashlaymiz, faqat haqiqiy tokenni qoldiramiz
		}

		// Tokenni tekshirish
		userID, err := utils.ValidateJWT(tokenStr)
		if err != nil {
			log.Printf("Invalid token: %v", err)
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		// Contextga userID ni qo‘shamiz
		ctx = context.WithValue(ctx, "userID", userID)
	}

	return handler(ctx, req)
}
