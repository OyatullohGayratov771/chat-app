package utils

import (
	"net"
	"net/http"
	"strings"
)

// GetIP — foydalanuvchi IP manzilini olish
func GetIP(r *http.Request) string {
	// Proxylardan o‘tgan bo‘lsa
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	ip = r.Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// GetUserAgent — foydalanuvchi qurilma haqida ma’lumot
func GetUserAgent(r *http.Request) string {
	return r.UserAgent()
}
