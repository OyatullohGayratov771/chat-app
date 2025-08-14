package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

type GeoIPResponse struct {
	Country string `json:"country"`
	Region  string `json:"regionName"`
	City    string `json:"city"`
	Query   string `json:"query"` // IP manzil
}

var ErrUnauthenticated = errors.New("unauthenticated")

func getIPFromCtx(ctx context.Context) *string {
	// 1. X-Forwarded-For dan olish
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get("x-forwarded-for"); len(vals) > 0 {
			parts := strings.Split(vals[0], ",")
			ip := strings.TrimSpace(parts[0])
			return &ip
		}
	}

	// 2. Agar topilmasa, peer (TCP connection) orqali olish
	if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
		ip := p.Addr.String()
		// IP:Port ko‘rinishida bo‘ladi, faqat IP qismini olish
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
		return &ip
	}

	return nil
}

func getUserAgentFromCtx(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if vals := md.Get("user-agent"); len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func GetLocationFromIP(ip string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://ip-api.com/json/%s", ip))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var data GeoIPResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	// Faqat country + city qaytaramiz
	return fmt.Sprintf("%s, %s", data.Country, data.City), nil
}
