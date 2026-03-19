package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"wack-backend/internal/config"
	"wack-backend/internal/model"
)

type SessionService struct {
	client *redis.Client
}

type SessionPayload struct {
	TokenID    string
	UserID     uint64
	Account    string
	Role       int
	Status     int
	DeviceType string
	IssuedAt   time.Time
	ExpiresAt  time.Time
}

func NewSessionService(cfg config.Config) (*SessionService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &SessionService{client: client}, nil
}

func (s *SessionService) CreateSession(ctx context.Context, payload SessionPayload) error {
	sessionKey := s.sessionKey(payload.TokenID)
	userSessionsKey := s.userSessionsKey(payload.UserID)
	ttl := payload.ExpiresAt.Sub(time.Now())
	if ttl <= 0 {
		ttl = time.Second
	}

	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, sessionKey,
		"user_id", strconv.FormatUint(payload.UserID, 10),
		"account", payload.Account,
		"role", strconv.Itoa(payload.Role),
		"status", strconv.Itoa(payload.Status),
		"device_type", payload.DeviceType,
		"issued_at", payload.IssuedAt.Format(time.RFC3339),
		"expires_at", payload.ExpiresAt.Format(time.RFC3339),
	)
	pipe.Expire(ctx, sessionKey, ttl)
	pipe.SAdd(ctx, userSessionsKey, payload.TokenID)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *SessionService) HasSession(ctx context.Context, tokenID string) (bool, error) {
	count, err := s.client.Exists(ctx, s.sessionKey(tokenID)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *SessionService) DeleteSession(ctx context.Context, tokenID string, userID uint64) error {
	pipe := s.client.TxPipeline()
	pipe.Del(ctx, s.sessionKey(tokenID))
	if userID != 0 {
		pipe.SRem(ctx, s.userSessionsKey(userID), tokenID)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *SessionService) DeleteAllUserSessions(ctx context.Context, userID uint64) error {
	userSessionsKey := s.userSessionsKey(userID)
	tokenIDs, err := s.client.SMembers(ctx, userSessionsKey).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	pipe := s.client.TxPipeline()
	if len(tokenIDs) > 0 {
		keys := make([]string, 0, len(tokenIDs))
		for _, tokenID := range tokenIDs {
			keys = append(keys, s.sessionKey(tokenID))
		}
		pipe.Del(ctx, keys...)
	}
	pipe.Del(ctx, userSessionsKey)
	_, err = pipe.Exec(ctx)
	return err
}

func (s *SessionService) Close() error {
	return s.client.Close()
}

func (s *SessionService) sessionKey(tokenID string) string {
	return "auth:session:" + tokenID
}

func (s *SessionService) userSessionsKey(userID uint64) string {
	return "auth:user_sessions:" + strconv.FormatUint(userID, 10)
}

func DeviceTypeForRole(role int) string {
	if role == model.RoleAdmin {
		return "admin"
	}
	return "mobile"
}

func SessionTTLForRole(role int) time.Duration {
	if role == model.RoleAdmin {
		return 7 * 24 * time.Hour
	}
	return 15 * 24 * time.Hour
}
