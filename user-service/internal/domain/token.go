package domain

// TokenProvider — JWT yoki boshqa token generatsiya qiluvchi abstraksiya
type TokenProvider interface {
	GenerateTokens(userID string) (string, string, error)
	RevokeRefreshToken(tokenStr string) error 
	ValidateAccessToken(tokenStr string) (userID string, err error)
	ValidateRefreshToken(tokenStr string) (string, error)
}
