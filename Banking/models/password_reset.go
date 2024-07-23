package models

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"myapp/rand"
	"strings"
	"time"
)

const DefaultResetDuration = 1 * time.Hour

type PasswordReset struct {
	ID int
	UserID int
	Token string
	TokenHash string
	ExpiresAt time.Time
}

type PasswordResetService struct {
	DB *sql.DB
	TokenBytes int
	ResetDuration time.Duration
}

func (prs *PasswordResetService) Create(email string) (*PasswordReset, error) {
	var psw PasswordReset
	email = strings.ToLower(email)
	row := prs.DB.QueryRow(`
		SELECT id FROM users WHERE email=$1;
	`, email)
	err := row.Scan(&psw.UserID)
	if err != nil {
		return nil, fmt.Errorf("password reset creation error: %w", err)
	}

	tokenBytes := prs.TokenBytes
	if tokenBytes < SessionTokenBytes {
		tokenBytes = SessionTokenBytes
	}
	psw.Token, err = rand.String(tokenBytes)
	if err != nil {
		return nil, fmt.Errorf("password reset creation error: %w", err)
	}
	psw.TokenHash = prs.hash(psw.Token)
	duration := prs.ResetDuration
	if duration < DefaultResetDuration {
		duration = DefaultResetDuration
	}
	psw.ExpiresAt = time.Now().Add(duration)
	row = prs.DB.QueryRow(`
		INSERT INTO password_resets(user_id, token_hash, expires_at)
		VALUES($1, $2, $3) ON CONFLICT (user_id) DO
		UPDATE
		SET token_hash=$2, expires_at=$3
		RETURNING id;
	`, psw.UserID, psw.TokenHash, psw.ExpiresAt)
	err = row.Scan(&psw.ID)
	if err != nil {
		return nil, fmt.Errorf("password reset creation error: %w", err)
	}
	return &psw, nil
}

func (prs *PasswordResetService) Consume(token string) (*User, error) {
	tokenHash := prs.hash(token)
	var user User
	var psw PasswordReset
	row := prs.DB.QueryRow(`
		SELECT password_resets.id, password_resets.expires_at,
		users.id, users.email, users.password_hash
		FROM password_resets
			JOIN users ON users.id = password_resets.user_id
		WHERE password_resets.token_hash=$1;
	`, tokenHash)
	err := row.Scan(
		&psw.ID, &psw.ExpiresAt,
		&user.ID, &user.Email, &user.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}
	if time.Now().After(psw.ExpiresAt) {
		return nil, fmt.Errorf("consume: %w", err)
	}
	err = prs.delete(psw.ID)
	if err != nil {
		return nil, fmt.Errorf("consume: %w", err)
	}
	return &user, nil
}

func (prs *PasswordResetService) hash(token string) string {
	tk := sha256.Sum256([]byte(token))
	return base64.URLEncoding.EncodeToString(tk[:])
}

func (prs *PasswordResetService) delete(id int) error {
	_, err := prs.DB.Exec(`
		DELETE FROM password_resets WHERE id=$1;
	`, id)
	if err != nil {
		return fmt.Errorf("prs deletion: %w", err)
	}
	return nil
}