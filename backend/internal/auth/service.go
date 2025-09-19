package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
    ErrUsernameTaken   = errors.New("username already taken")
    ErrInvalidLogin    = errors.New("invalid username or password")
    ErrWeakPassword    = errors.New("password too weak (min 6 chars)")
)

type AuthService struct {
    users repository.UserRepository
    jwt   *JWTManager
}

func NewAuthService(users repository.UserRepository, jwt *JWTManager) *AuthService {
    return &AuthService{users: users, jwt: jwt}
}

func (s *AuthService) Register(ctx context.Context, username, password string, roles []string) (domain.User, TokenPair, error) {
    username = strings.TrimSpace(username)
    if len(username) < 3 { return domain.User{}, TokenPair{}, errors.New("username too short") }
    if len(password) < 6 { return domain.User{}, TokenPair{}, ErrWeakPassword }

    // hash
    hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil { return domain.User{}, TokenPair{}, err }

    u := domain.User{Username: username, Roles: roles, PasswordHash: string(hashBytes), CreatedAt: time.Now().UTC()}
    if err := s.users.Create(ctx, u); err != nil {
        if errors.Is(err, repository.ErrUserDuplicate) { return domain.User{}, TokenPair{}, ErrUsernameTaken }
        return domain.User{}, TokenPair{}, err
    }
    // issue tokens
    access, expA, err := s.jwt.GenerateAccess(u.ID, u.Roles)
    if err != nil { return domain.User{}, TokenPair{}, err }
    refresh, expR, err := s.jwt.GenerateRefresh(u.ID)
    if err != nil { return domain.User{}, TokenPair{}, err }
    pair := TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(expA.Sub(time.Now().UTC()).Seconds())}
    _ = expR // 可在后续返回 refresh 过期信息
    return u, pair, nil
}

func (s *AuthService) Authenticate(ctx context.Context, username, password string) (domain.User, TokenPair, error) {
    u, err := s.users.GetByUsername(ctx, username)
    if err != nil { return domain.User{}, TokenPair{}, ErrInvalidLogin }
    if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil { return domain.User{}, TokenPair{}, ErrInvalidLogin }
    access, expA, err := s.jwt.GenerateAccess(u.ID, u.Roles)
    if err != nil { return domain.User{}, TokenPair{}, err }
    refresh, _, err := s.jwt.GenerateRefresh(u.ID)
    if err != nil { return domain.User{}, TokenPair{}, err }
    pair := TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(expA.Sub(time.Now().UTC()).Seconds())}
    return u, pair, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (TokenPair, error) {
    claims, err := s.jwt.ParseRefresh(refreshToken)
    if err != nil { return TokenPair{}, ErrInvalidToken }
    u, err := s.users.GetByID(ctx, claims.UserID)
    if err != nil { return TokenPair{}, ErrInvalidToken }
    access, expA, err := s.jwt.GenerateAccess(u.ID, u.Roles)
    if err != nil { return TokenPair{}, err }
    refresh, _, err := s.jwt.GenerateRefresh(u.ID)
    if err != nil { return TokenPair{}, err }
    return TokenPair{AccessToken: access, RefreshToken: refresh, ExpiresIn: int64(expA.Sub(time.Now().UTC()).Seconds())}, nil
}
