package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"simvex/internal/models"
	"simvex/internal/repository"
)

const (
	verifyPrefix  = "verify:"
	resetPrefix   = "reset:"
	sessionPrefix = "session:"
	userSessPref  = "user_sessions:"
)

type AuthService struct {
	repo       repository.Repository
	redis      *redis.Client
	mailer     Mailer
	sessionTTL time.Duration
	codeTTL    time.Duration
}

func NewAuthService(repo repository.Repository, redisClient *redis.Client, mailer Mailer, sessionTTL, codeTTL time.Duration) *AuthService {
	return &AuthService{
		repo:       repo,
		redis:      redisClient,
		mailer:     mailer,
		sessionTTL: sessionTTL,
		codeTTL:    codeTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(password) < 8 {
		return ErrInvalidCredentials
	}

	existing, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	if existing == nil {
		if _, err := s.repo.CreateUser(email, string(hash)); err != nil {
			return err
		}
	} else {
		if existing.IsVerified {
			return ErrUserAlreadyExists
		}
		if err := s.repo.UpdateUserPassword(existing.ID, string(hash)); err != nil {
			return err
		}
	}

	code, err := randomCode(6)
	if err != nil {
		return err
	}
	if err := s.redis.Set(ctx, verifyPrefix+email, code, s.codeTTL).Err(); err != nil {
		return err
	}

	if s.mailer == nil {
		return fmt.Errorf("mailer not configured")
	}
	subject := "SIMVEX 이메일 인증 코드"
	body := buildAuthEmailHTML("이메일 인증", code)
	return s.mailer.Send(email, subject, body, true)
}

func (s *AuthService) VerifyEmail(ctx context.Context, email, code string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	stored, err := s.redis.Get(ctx, verifyPrefix+email).Result()
	if err == redis.Nil {
		return ErrCodeExpired
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(code) != stored {
		return ErrCodeMismatch
	}

	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}
	if err := s.repo.SetUserVerified(user.ID); err != nil {
		return err
	}
	_ = s.redis.Del(ctx, verifyPrefix+email).Err()
	return nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrUserNotFound
	}
	if !user.IsVerified {
		return "", ErrUserNotVerified
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", ErrInvalidCredentials
	}
	return s.createSession(ctx, user.ID)
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	userID, err := s.redis.Get(ctx, sessionPrefix+sessionID).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	if userID != "" {
		_ = s.redis.SRem(ctx, userSessPref+userID, sessionID).Err()
	}
	_ = s.redis.Del(ctx, sessionPrefix+sessionID).Err()
	return nil
}

func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}
	code, err := randomCode(6)
	if err != nil {
		return err
	}
	if err := s.redis.Set(ctx, resetPrefix+email, code, s.codeTTL).Err(); err != nil {
		return err
	}
	if s.mailer == nil {
		return fmt.Errorf("mailer not configured")
	}
	subject := "SIMVEX 비밀번호 재설정 코드"
	body := buildAuthEmailHTML("비밀번호 재설정", code)
	return s.mailer.Send(email, subject, body, true)
}

func (s *AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	stored, err := s.redis.Get(ctx, resetPrefix+email).Result()
	if err == redis.Nil {
		return ErrCodeExpired
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(code) != stored {
		return ErrCodeMismatch
	}
	if len(newPassword) < 8 {
		return ErrInvalidCredentials
	}

	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdateUserPassword(user.ID, string(hash)); err != nil {
		return err
	}
	_ = s.redis.Del(ctx, resetPrefix+email).Err()
	_ = s.invalidateAllSessions(ctx, user.ID)
	return nil
}

func (s *AuthService) createSession(ctx context.Context, userID string) (string, error) {
	sessionID, err := randomToken(32)
	if err != nil {
		return "", err
	}
	if err := s.redis.Set(ctx, sessionPrefix+sessionID, userID, s.sessionTTL).Err(); err != nil {
		return "", err
	}
	if err := s.redis.SAdd(ctx, userSessPref+userID, sessionID).Err(); err != nil {
		return "", err
	}
	_ = s.redis.Expire(ctx, userSessPref+userID, s.sessionTTL).Err()
	return sessionID, nil
}

func (s *AuthService) invalidateAllSessions(ctx context.Context, userID string) error {
	sessions, err := s.redis.SMembers(ctx, userSessPref+userID).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	for _, sid := range sessions {
		_ = s.redis.Del(ctx, sessionPrefix+sid).Err()
	}
	_ = s.redis.Del(ctx, userSessPref+userID).Err()
	return nil
}

func randomToken(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func randomCode(length int) (string, error) {
	token, err := randomToken(length)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(token)[:length], nil
}

var (
	ErrInvalidCredentials = fmt.Errorf("invalid credentials")
	ErrUserAlreadyExists  = fmt.Errorf("user already exists")
	ErrUserNotFound       = fmt.Errorf("user not found")
	ErrUserNotVerified    = fmt.Errorf("user not verified")
	ErrCodeMismatch       = fmt.Errorf("code mismatch")
	ErrCodeExpired        = fmt.Errorf("code expired")
)

func (s *AuthService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.repo.GetUserByID(id)
}

func buildAuthEmailHTML(title, code string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="ko">
  <body style="margin:0;padding:0;background:#0d1117;font-family:Arial,sans-serif;color:#f5f3ee;">
    <div style="max-width:560px;margin:40px auto;background:#121c2b;border-radius:16px;padding:24px;border:1px solid rgba(255,255,255,0.08);">
      <h2 style="margin:0 0 12px 0;color:#f8c86a;">SIMVEX %s</h2>
      <p style="margin:0 0 16px 0;color:#cbd5e1;">아래 인증 코드를 10분 안에 입력해주세요.</p>
      <div style="font-size:28px;font-weight:bold;letter-spacing:4px;color:#ffffff;background:#0c1118;border-radius:12px;padding:16px;text-align:center;">
        %s
      </div>
      <p style="margin:16px 0 0 0;color:#94a3b8;font-size:12px;">이 요청을 본인이 하지 않았다면 이 이메일을 무시하세요.</p>
    </div>
  </body>
</html>`, title, code)
}
