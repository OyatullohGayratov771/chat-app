package domain

import (
	"context"
	"time"
)

// ======================
// ENTITY
// ======================
type User struct {
	ID            string
	Username      *string
	Email         *string
	PasswordHash  string
	FullName      *string
	AvatarURL     *string
	Language      *string
	Platform      string
	DeviceID      string
	RegisteredIP  *string
	UserAgent     string
	Location      *string
	RegisteredAt  time.Time
	UpdatedAt     time.Time
}

// ======================
// REPOSITORY INTERFACE
// ======================
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)

	// Update one specific field
	UpdateField(ctx context.Context, userID string, field string, value *string) error

	ChangePassword(ctx context.Context, id, newHash string) error
	ResetPassword(ctx context.Context, id, newHash string) error
	Delete(ctx context.Context, id string) error

	// Session management
	GetSessions(ctx context.Context, userID string) ([]Session, error)
	DeleteSession(ctx context.Context, userID, deviceID string) error
	DeleteAllSessions(ctx context.Context, userID string) error
}

// ======================
// SERVICE INTERFACE
// ======================
type UserService interface {
	// Auth
	Register(ctx context.Context, req RegisterDTO) (*AuthResult, error)
	Login(ctx context.Context, email, password, platform, deviceID string) (*AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	Logout(ctx context.Context, userID, deviceID string) error

	// Profile
	GetProfile(ctx context.Context, userID string) (*User, error)
	UpdateUsername(ctx context.Context, userID, username string) (*User, error)
	UpdateEmail(ctx context.Context, userID, email string) (*User, error)
	UpdateFullName(ctx context.Context, userID, fullName string) (*User, error)
	UpdateAvatar(ctx context.Context, userID, avatarURL string) (*User, error)
	UpdateLanguage(ctx context.Context, userID, language string) (*User, error)

	// Security
	ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error

	// Account
	DeleteAccount(ctx context.Context, userID string) error

	// Sessions
	GetSessions(ctx context.Context, userID string) ([]Session, error)
}

// ======================
// DTOs
// ======================
type RegisterDTO struct {
	Username     string
	Email        string
	Password     string
	FullName     *string
	AvatarURL    *string
	Language     *string
	Platform     string
	DeviceID     string
	RegisteredIP *string
	UserAgent    string
	Location     *string
}

// ======================
// SESSION
// ======================
type Session struct {
	DeviceID  string
	Platform  string
	IPAddress string
	LastSeen  time.Time
}

// ======================
// AUTH RESULT
// ======================
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         *User
}
