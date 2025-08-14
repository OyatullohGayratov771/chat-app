package main

import (
	"context"
	"log"
	"net"
	"strings"
	"time"

	"user-service/internal/cache/redis"
	"user-service/internal/config"
	"user-service/internal/domain"
	"user-service/internal/event/kafka"
	grpcserver "user-service/internal/handler/grpc"
	"user-service/internal/repository/postgres"
	service "user-service/internal/service/user"
	"user-service/internal/storage"
	"user-service/internal/utils"

	pb "user-service/protos/user"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

var ErrUnauthenticated = status.Error(codes.Unauthenticated, "unauthenticated")

func main() {
	// 1. Config yuklash
	config.LoadConfig()
	cfg := config.AppConfig

	// 2. PostgreSQL ulanish
	db, err := storage.ConnectToDB(cfg)
	if err != nil {
		log.Fatalf("❌ Failed to connect to PostgreSQL: %v", err)
	}
	defer db.Close()

	// 3. Kafka producer
	kafkaProducer := kafka.NewKafkaProducer(
		[]string{cfg.Kafka.Host + ":" + cfg.Kafka.Port},
		cfg.Kafka.Topic,
	)

	// 4. Redis client
	redisClient := redis.NewRedisClient(cfg)

	// 5. JWT Provider
	tokenProvider := utils.NewJWTProvider(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		time.Minute*15,
		time.Hour*24*7,
		redisClient,
	)

	// 6. Repository
	userRepo := postgres.NewUserRepository(db)

	// 7. Service layer
	userService := service.NewUserService(userRepo, tokenProvider, kafkaProducer)

	// 8. gRPC server + Auth interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(authInterceptor(tokenProvider)),
	)
	pb.RegisterUserServiceServer(grpcServer, grpcserver.NewUserServer(userService))

	// 9. Reflection (grpcurl uchun)
	reflection.Register(grpcServer)

	// 10. TCP listener
	addr := cfg.Http.Host + ":" + cfg.Http.Port
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("❌ Failed to listen on %s: %v", addr, err)
	}

	log.Printf("✅ gRPC User Service is running at %s", addr)

	// 11. Serverni ishga tushirish
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Failed to serve gRPC server: %v", err)
	}
}

// Auth interceptor — barcha RPC chaqiriqlarda tokenni tekshiradi
func authInterceptor(tokenProvider domain.TokenProvider) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// Login va Register endpointlarini token tekshirishdan chiqaramiz
		if strings.Contains(info.FullMethod, "Login") ||
			strings.Contains(info.FullMethod, "Register") {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrUnauthenticated
		}

		tokens := md["authorization"]
		if len(tokens) == 0 {
			return nil, ErrUnauthenticated
		}

		userID, err := tokenProvider.ValidateAccessToken(strings.TrimPrefix(tokens[0], "Bearer "))
		if err != nil {
			return nil, ErrUnauthenticated
		}
		// Typed context key ishlatamiz
		return handler(context.WithValue(ctx, "userID", userID), req)
	}
}
