package grpc

import (
	"context"
	"time"

	"user-service/internal/domain"
	userpb "user-service/protos/user"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserServer struct {
	userpb.UnimplementedUserServiceServer
	userService domain.UserService
}

func NewUserServer(userService domain.UserService) *UserServer {
	return &UserServer{
		userService: userService,
	}
}

// =====================
// REGISTER
// =====================
func (s *UserServer) Register(ctx context.Context, req *userpb.RegisterRequest) (*userpb.AuthResponse, error) {
	ip := getIPFromCtx(ctx)
	loc, _ := GetLocationFromIP(*ip)

	dto := domain.RegisterDTO{
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		FullName:     toPtr(req.FullName),
		AvatarURL:    toPtr(req.AvatarUrl),
		Language:     toPtr(req.Language),
		Platform:     req.Platform,
		DeviceID:     req.DeviceId,
		RegisteredIP: ip,
		UserAgent:    getUserAgentFromCtx(ctx),
		Location:     toPtr(loc),
	}

	authResult, err := s.userService.Register(ctx, dto)
	if err != nil {
		return nil, err
	}

	return toAuthResponse(authResult), nil
}

// =====================
// LOGIN
// =====================
func (s *UserServer) Login(ctx context.Context, req *userpb.LoginRequest) (*userpb.AuthResponse, error) {
	authResult, err := s.userService.Login(ctx, req.Email, req.Password, req.Platform, req.DeviceId)
	if err != nil {
		return nil, err
	}
	return toAuthResponse(authResult), nil
}

// =====================
// GET PROFILE
// =====================
func (s *UserServer) GetProfile(ctx context.Context, _ *userpb.Empty) (*userpb.User, error) {

	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	user, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	return toUserPB(user), nil
}

// =====================
// LOGOUT
// =====================
func (s *UserServer) Logout(ctx context.Context, _ *userpb.Empty) (*userpb.Empty, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	user, err := s.userService.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := s.userService.Logout(ctx, userID, user.DeviceID); err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

// =====================
// DELETE ACCOUNT
// =====================
func (s *UserServer) DeleteAccount(ctx context.Context, _ *userpb.Empty) (*userpb.Empty, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	if err := s.userService.DeleteAccount(ctx, userID); err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

// =====================
// GET SESSIONS
// =====================
func (s *UserServer) GetSessions(ctx context.Context, _ *userpb.Empty) (*userpb.SessionList, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	sessions, err := s.userService.GetSessions(ctx, userID)
	if err != nil {
		return nil, err
	}
	Sessions := make([]*userpb.Session, len(sessions))
	for i, s := range sessions {
		Sessions[i] = &userpb.Session{
			DeviceId:     s.DeviceID,
			Platform:     s.Platform,
			IpAddress:    s.IPAddress,
			LastSeen:     toProtoTime(s.LastSeen),
		}
	}
	return &userpb.SessionList{Sessions: Sessions}, nil
}


// =====================
// REFRESH TOKEN
// =====================
func (s *UserServer) RefreshToken(ctx context.Context, req *userpb.RefreshTokenRequest) (*userpb.AuthResponse, error) {
	authResult, err := s.userService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}
	return toAuthResponse(authResult), nil
}

// =====================
// update username
// =====================
func (s *UserServer) UpdateUsername(ctx context.Context, req *userpb.UpdateUsernameRequest) (*userpb.User, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	user, err := s.userService.UpdateUsername(ctx, userID, req.Username)
	if err != nil {
		return nil, err
	}
	return toUserPB(user), nil
}

// =====================
// update email
// =====================
func (s *UserServer) UpdateEmail(ctx context.Context, req *userpb.UpdateEmailRequest) (*userpb.User, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	user, err := s.userService.UpdateEmail(ctx, userID, req.Email)
	if err != nil {
		return nil, err
	}
	return toUserPB(user), nil
}


// =====================
// update full name
// =====================
func (s *UserServer) UpdateFullName(ctx context.Context, req *userpb.UpdateFullNameRequest) (*userpb.User, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	user, err := s.userService.UpdateFullName(ctx, userID, req.FullName)
	if err != nil {
		return nil, err
	}
	return toUserPB(user), nil
}


// =====================
// update avatar url
// =====================
func (s *UserServer) UpdateAvatar(ctx context.Context, req *userpb.UpdateAvatarRequest) (*userpb.User, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	user, err := s.userService.UpdateAvatar(ctx, userID, req.AvatarUrl)
	if err != nil {
		return nil, err
	}
	return toUserPB(user), nil
}

// =====================
// update language
// =====================
func (s *UserServer) UpdateLanguage(ctx context.Context, req *userpb.UpdateLanguageRequest) (*userpb.User, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	user, err := s.userService.UpdateLanguage(ctx, userID, req.Language)
	if err != nil {
		return nil, err
	}
	return toUserPB(user), nil
}


// =====================
// change password
// =====================
func (s *UserServer) ChangePassword(ctx context.Context, req *userpb.ChangePasswordRequest) (*userpb.Empty, error) {
	userID, ok := ctx.Value("userID").(string)
	if !ok || userID == "" {
		return nil, ErrUnauthenticated
	}
	if err := s.userService.ChangePassword(ctx, userID, req.OldPassword, req.NewPassword); err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

// =====================
// forgot password
// =====================
func (s *UserServer) ForgotPassword(ctx context.Context, req *userpb.ForgotPasswordRequest) (*userpb.Empty, error) {
	if err := s.userService.ForgotPassword(ctx, req.Email); err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}


// =====================
// reset password
// =====================
func (s *UserServer) ResetPassword(ctx context.Context, req *userpb.ResetPasswordRequest) (*userpb.Empty, error) {
	if err := s.userService.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		return nil, err
	}
	return &userpb.Empty{}, nil
}

// =====================
// HELPERS
// =====================
func toUserPB(u *domain.User) *userpb.User {
	if u == nil {
		return nil
	}

	return &userpb.User{
		Id:           u.ID,
		Username:     getStr(u.Username),
		Email:        getStr(u.Email),
		FullName:     getStr(u.FullName),
		AvatarUrl:    getStr(u.AvatarURL),
		Language:     getStr(u.Language),
		Platform:     u.Platform,
		DeviceId:     u.DeviceID,
		RegisteredIp: getStr(u.RegisteredIP),
		UserAgent:    u.UserAgent,
		Location:     getStr(u.Location),
		RegisteredAt: toProtoTime(u.RegisteredAt),
		UpdatedAt:    toProtoTime(u.UpdatedAt),
	}
}

func toAuthResponse(a *domain.AuthResult) *userpb.AuthResponse {
	return &userpb.AuthResponse{
		AccessToken:  a.AccessToken,
		RefreshToken: a.RefreshToken,
		User:         toUserPB(a.User),
	}
}

// pointer helpers
func toPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func getStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func toProtoTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
