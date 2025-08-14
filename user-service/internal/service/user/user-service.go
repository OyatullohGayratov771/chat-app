package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"
	"user-service/internal/domain"
	"user-service/internal/event/kafka"

	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo          domain.UserRepository
	tokenProvider domain.TokenProvider
	k             kafka.KafkaProducer
}

func NewUserService(repo domain.UserRepository, tokenProvider domain.TokenProvider, kafka *kafka.KafkaProducer) domain.UserService {
	return &userService{
		repo:          repo,
		tokenProvider: tokenProvider,
		k:             *kafka,
	}
}

// ================= REGISTER =================
func (s *userService) Register(ctx context.Context, req domain.RegisterDTO) (*domain.AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))
	if existing, _ := s.repo.GetByEmail(ctx, email); existing != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &domain.User{
		Username:     &req.Username,
		Email:        &email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
		AvatarURL:    req.AvatarURL,
		Language:     req.Language,
		Platform:     req.Platform,
		DeviceID:     req.DeviceID,
		RegisteredIP: req.RegisteredIP,
		UserAgent:    req.UserAgent,
		Location:     req.Location,
		RegisteredAt: time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	event := map[string]string{
		"event":    "UserRegistered",
		"user_id":  user.ID,
		"email":    *user.Email,
		"username": *user.Username,
	}
	if eventBytes, _ := json.Marshal(event); true {
		if err := s.k.Publish(ctx, eventBytes); err != nil {
			log.Println("Kafka publish error:", err)
		}
	}

	access, refresh, err := s.tokenProvider.GenerateTokens(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	return &domain.AuthResult{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         user,
	}, nil
}

// ================= LOGIN =================
func (s *userService) Login(ctx context.Context, email, password, platform, deviceID string) (*domain.AuthResult, error) {
	user, err := s.repo.GetByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if err != nil || user == nil {
		return nil, errors.New("invalid email or password")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid email or password")
	}

	access, refresh, err := s.tokenProvider.GenerateTokens(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	return &domain.AuthResult{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         user,
	}, nil
}

// ================= GET PROFILE =================
func (s *userService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	return s.repo.GetByID(ctx, userID)
}

// ================= UPDATE FIELDS =================
func (s *userService) UpdateUsername(ctx context.Context, userID, username string) (*domain.User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}
	if err := s.repo.UpdateField(ctx, userID, "username", &username); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID)
}

func (s *userService) UpdateEmail(ctx context.Context, userID, email string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if err := s.repo.UpdateField(ctx, userID, "email", &email); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID)
}

func (s *userService) UpdateFullName(ctx context.Context, userID, fullName string) (*domain.User, error) {
	if fullName == "" {
		return nil, errors.New("full_name cannot be empty")
	}
	if err := s.repo.UpdateField(ctx, userID, "full_name", &fullName); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID)
}

func (s *userService) UpdateAvatar(ctx context.Context, userID, avatarURL string) (*domain.User, error) {
	var value *string
	if strings.TrimSpace(avatarURL) != "" {
		value = &avatarURL
	} // else -> nil = remove avatar
	if err := s.repo.UpdateField(ctx, userID, "avatar_url", value); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID)
}

func (s *userService) UpdateLanguage(ctx context.Context, userID, language string) (*domain.User, error) {
	if language == "" {
		return nil, errors.New("language cannot be empty")
	}
	if err := s.repo.UpdateField(ctx, userID, "language", &language); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, userID)
}

// ================= REFRESH TOKEN =================
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResult, error) {
	userID, err := s.tokenProvider.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	access, refresh, err := s.tokenProvider.GenerateTokens(user.ID)
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	return &domain.AuthResult{
		AccessToken:  access,
		RefreshToken: refresh,
		User:         user,
	}, nil
}

// ================= LOGOUT =================
func (s *userService) Logout(ctx context.Context, userID, deviceID string) error {
	return s.repo.DeleteSession(ctx, userID, deviceID)
}

// ================= CHANGE PASSWORD =================
func (s *userService) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)) != nil {
		return errors.New("old password is incorrect")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	return s.repo.ChangePassword(ctx, userID, string(newHash))
}

// ================= FORGOT PASSWORD =================
func (s *userService) ForgotPassword(ctx context.Context, email string) error {
	return nil
}

// ================= RESET PASSWORD =================
func (s *userService) ResetPassword(ctx context.Context, token, newPassword string) error {
	
	return s.repo.ResetPassword(ctx, "","" )
}

// ================= DELETE ACCOUNT =================
func (s *userService) DeleteAccount(ctx context.Context, userID string) error {
	return s.repo.Delete(ctx, userID)
}

// ================= GET SESSIONS =================
func (s *userService) GetSessions(ctx context.Context, userID string) ([]domain.Session, error) {
	return s.repo.GetSessions(ctx, userID)
}

// ================= DELETE SESSION =================
func (s *userService) DeleteSession(ctx context.Context, userID, deviceID string) error {
	return s.repo.DeleteSession(ctx, userID, deviceID)
}

// ================= DELETE ALL SESSIONS =================
func (s *userService) DeleteAllSessions(ctx context.Context, userID string) error {
	return s.repo.DeleteAllSessions(ctx, userID)
}
