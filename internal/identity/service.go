package identity

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/faria/traveling-hub/internal/frog"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type Service struct {
	db              *sql.DB
	sessions        sessionStore
	autoVerifyEmail bool
	mailer          VerificationMailer
	webOrigin       string
}

func NewService(db *sql.DB, redisClient *redis.Client, sessionTTL time.Duration, autoVerifyEmail bool, mailer VerificationMailer, webOrigin string) Service {
	return Service{
		db: db, sessions: newSessionStore(redisClient, sessionTTL), autoVerifyEmail: autoVerifyEmail, mailer: mailer, webOrigin: webOrigin,
	}
}

func (s Service) Register(ctx context.Context, email string) (RegistrationResult, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return RegistrationResult{}, err
	}
	credentials, err := NewCredentials()
	if err != nil {
		return RegistrationResult{}, err
	}
	passwordHash, err := HashPassword(credentials.InitialPassword)
	if err != nil {
		return RegistrationResult{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return RegistrationResult{}, fmt.Errorf("begin registration: %w", err)
	}
	defer tx.Rollback()

	userID, agentID := uuid.New(), uuid.New()
	verificationToken := ""
	var verifiedAt any
	if s.autoVerifyEmail {
		verifiedAt = time.Now().UTC()
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (id, email_normalized, password_hash, must_change_password, email_verified_at)
		VALUES ($1, $2, $3, TRUE, $4)`, userID, email, passwordHash, verifiedAt); err != nil {
		if isUniqueViolation(err) {
			return RegistrationResult{Created: false}, nil
		}
		return RegistrationResult{}, fmt.Errorf("insert user: %w", err)
	}
	if !s.autoVerifyEmail {
		verificationToken, err = NewVerificationToken()
		if err != nil {
			return RegistrationResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO email_verification_tokens (token_digest, user_id, expires_at) VALUES ($1, $2, NOW() + INTERVAL '24 hours')`, Digest(verificationToken), userID); err != nil {
			return RegistrationResult{}, fmt.Errorf("insert email verification token: %w", err)
		}
	}
	keyDigest := Digest(credentials.AgentAPIKey)
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO agents (id, user_id, token_digest, api_key_digest)
		VALUES ($1, $2, $3, $3)`, agentID, userID, keyDigest); err != nil {
		return RegistrationResult{}, fmt.Errorf("insert agent: %w", err)
	}
	f, err := frog.InsertForAgentTx(ctx, tx, agentID)
	if err != nil {
		return RegistrationResult{}, err
	}
	if err := tx.Commit(); err != nil {
		return RegistrationResult{}, fmt.Errorf("commit registration: %w", err)
	}
	if verificationToken != "" && s.mailer != nil {
		url, err := verificationURL(s.webOrigin, verificationToken)
		if err != nil {
			return RegistrationResult{}, err
		}
		if err := s.mailer.SendVerification(ctx, email, url); err != nil {
			return RegistrationResult{}, fmt.Errorf("send verification email: %w", err)
		}
	}
	return RegistrationResult{
		Created: true, AgentID: agentID, FrogID: f.ID, Username: email,
		InitialPassword: credentials.InitialPassword, AgentAPIKey: credentials.AgentAPIKey, MustChange: true,
	}, nil
}

func (s Service) Login(ctx context.Context, email, password string) (string, UserSession, error) {
	email, err := NormalizeEmail(email)
	if err != nil {
		return "", UserSession{}, fmt.Errorf("invalid login")
	}
	var userID uuid.UUID
	var hash string
	var mustChange bool
	var sessionVersion int64
	var verifiedAt sql.NullTime
	if err := s.db.QueryRowContext(ctx, `
		SELECT id, password_hash, must_change_password, email_verified_at, session_version
		FROM users WHERE email_normalized = $1`, email).Scan(&userID, &hash, &mustChange, &verifiedAt, &sessionVersion); err != nil {
		return "", UserSession{}, fmt.Errorf("invalid login")
	}
	if !VerifyPassword(hash, password) || (!s.autoVerifyEmail && !verifiedAt.Valid) {
		return "", UserSession{}, fmt.Errorf("invalid login")
	}
	session := UserSession{UserID: userID, MustChange: mustChange, SessionVersion: sessionVersion}
	id, err := s.sessions.Create(ctx, session)
	if err != nil {
		return "", UserSession{}, err
	}
	return id, session, nil
}

// VerifyEmail consumes a one-time, opaque verification token. Tokens are
// stored only as digests and an expired or previously consumed token reveals
// nothing about the user it was issued for.
func (s Service) VerifyEmail(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("invalid verification token")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin email verification: %w", err)
	}
	defer tx.Rollback()
	var userID uuid.UUID
	err = tx.QueryRowContext(ctx, `
		UPDATE email_verification_tokens
		SET consumed_at = NOW()
		WHERE token_digest = $1 AND consumed_at IS NULL AND expires_at > NOW()
		RETURNING user_id`, Digest(token)).Scan(&userID)
	if err != nil {
		return fmt.Errorf("invalid verification token")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE users SET email_verified_at = COALESCE(email_verified_at, NOW()), updated_at = NOW() WHERE id = $1`, userID); err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit email verification: %w", err)
	}
	return nil
}

func (s Service) LoadSession(ctx context.Context, sessionID string) (UserSession, error) {
	session, err := s.sessions.Load(ctx, sessionID)
	if err != nil {
		return UserSession{}, err
	}
	var currentVersion int64
	if err := s.db.QueryRowContext(ctx, `SELECT session_version FROM users WHERE id = $1`, session.UserID).Scan(&currentVersion); err != nil {
		return UserSession{}, fmt.Errorf("load user session version: %w", err)
	}
	if session.SessionVersion != currentVersion {
		_ = s.sessions.Delete(ctx, sessionID)
		return UserSession{}, fmt.Errorf("stale session")
	}
	return session, nil
}

func (s Service) ChangePassword(ctx context.Context, sessionID string, session UserSession, password string) (string, error) {
	hash, err := HashPassword(password)
	if err != nil {
		return "", err
	}
	var nextVersion int64
	err = s.db.QueryRowContext(ctx, `
		UPDATE users
		SET password_hash = $1, must_change_password = FALSE, session_version = session_version + 1, updated_at = NOW()
		WHERE id = $2 AND session_version = $3
		RETURNING session_version`, hash, session.UserID, session.SessionVersion).Scan(&nextVersion)
	if err != nil {
		return "", fmt.Errorf("update password: %w", err)
	}
	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		return "", fmt.Errorf("delete restricted session: %w", err)
	}
	return s.sessions.Create(ctx, UserSession{UserID: session.UserID, MustChange: false, SessionVersion: nextVersion})
}

func NormalizeEmail(email string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(email))
	if len(normalized) > 320 || !emailPattern.MatchString(normalized) {
		return "", fmt.Errorf("invalid email")
	}
	return normalized, nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
