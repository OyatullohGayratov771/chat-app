package utils

// import (
// 	"context"
// 	"log"
// 	"strings"

// 	"google.golang.org/grpc"
// 	"google.golang.org/grpc/codes"
// 	"google.golang.org/grpc/metadata"
// 	"google.golang.org/grpc/status"
// )

// // Key to extract user ID from context later
// type ctxKey string

// const UserIDKey ctxKey = "userID"

// // gRPC Unary Interceptor for JWT authentication
// func GRPCAuthInterceptor(
// 	ctx context.Context,
// 	req interface{},
// 	info *grpc.UnaryServerInfo,
// 	handler grpc.UnaryHandler,
// ) (interface{}, error) {
// 	md, ok := metadata.FromIncomingContext(ctx)
// 	if !ok {
// 		return handler(ctx, req)
// 	}

// 	authHeaders := md.Get("authorization")
// 	if len(authHeaders) == 0 {
// 		return handler(ctx, req) // No token → public endpoint
// 	}

// 	tokenStr := strings.TrimPrefix(authHeaders[0], "Bearer ")

// 	userID, err := ValidateAccessToken(tokenStr)
// 	if err != nil {
// 		log.Printf("❌ Invalid token: %v", err)
// 		return nil, status.Error(codes.Unauthenticated, "invalid token")
// 	}

// 	ctx = context.WithValue(ctx, UserIDKey, userID)
// 	return handler(ctx, req)
// }

// // Helper to extract userID from context
// func GetUserIDFromContext(ctx context.Context) (string, bool) {
// 	userID, ok := ctx.Value(UserIDKey).(string)
// 	return userID, ok
// }
