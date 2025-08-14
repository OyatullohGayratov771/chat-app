package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"user-service/internal/domain"

	"github.com/google/uuid"
)

type userRepository struct {
	db *sql.DB
}

// Constructor
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &userRepository{db: db}
}

// ================== CREATE ==================
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (
			id, username, email, password, full_name, avatar_url, language,
			platform, device_id, registered_ip, user_agent, location,
			registered_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12,
			NOW(), NOW()
		)
	`
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.AvatarURL,
		user.Language,
		user.Platform,
		user.DeviceID,
		user.RegisteredIP,
		user.UserAgent,
		user.Location,
	)
	return err
}

// ================== GET BY ID ==================
func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, full_name, avatar_url, language,
		       platform, device_id, registered_ip, user_agent, location,
		       registered_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.Language,
		&user.Platform,
		&user.DeviceID,
		&user.RegisteredIP,
		&user.UserAgent,
		&user.Location,
		&user.RegisteredAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// ================== GET BY EMAIL ==================
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, username, email, password, full_name, avatar_url, language,
		       platform, device_id, registered_ip, user_agent, location,
		       registered_at, updated_at
		FROM users
		WHERE email = $1
	`
	row := r.db.QueryRowContext(ctx, query, email)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.AvatarURL,
		&user.Language,
		&user.Platform,
		&user.DeviceID,
		&user.RegisteredIP,
		&user.UserAgent,
		&user.Location,
		&user.RegisteredAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// ================== UPDATE ONE FIELD ==================
// value = nil  => SET field = NULL
func (r *userRepository) UpdateField(ctx context.Context, userID string, field string, value *string) error {
	// Whitelist field names to avoid SQL injection
	allowedFields := map[string]bool{
		"username":   true,
		"email":      true,
		"full_name":  true,
		"avatar_url": true,
		"language":   true,
	}

	if !allowedFields[field] {
		return fmt.Errorf("invalid field name: %s", field)
	}

	query := fmt.Sprintf(`UPDATE users SET %s = $1, updated_at = NOW() WHERE id = $2`, field)
	_, err := r.db.ExecContext(ctx, query, value, userID)
	return err
}

// ================== CHANGE PASSWORD ==================
func (r *userRepository) ChangePassword(ctx context.Context, id, newHash string) error {
	query := `
		UPDATE users
		SET password = $1,
		    updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, newHash, id)
	return err
}

func (r *userRepository) ResetPassword(ctx context.Context, id, newHash string) error{
	query := `
		UPDATE users
		SET password = $1,
		    updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.ExecContext(ctx, query, newHash, id)
	return err
}

// ================== DELETE ACCOUNT ==================
func (r *userRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	return err
}

// ================== SESSIONS ==================
func (r *userRepository) GetSessions(ctx context.Context, userID string) ([]domain.Session, error) {
	query := `
		SELECT device_id, platform, ip_address, last_seen
		FROM sessions
		WHERE user_id = $1
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session
		err := rows.Scan(&s.DeviceID, &s.Platform, &s.IPAddress, &s.LastSeen)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *userRepository) DeleteSession(ctx context.Context, userID, deviceID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1 AND device_id = $2`, userID, deviceID)
	return err
}

func (r *userRepository) DeleteAllSessions(ctx context.Context, userID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}
