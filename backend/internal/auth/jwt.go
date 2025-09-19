package auth

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"` // access token 剩余秒数
}

type JWTManager struct {
    accessTTL  time.Duration
    refreshTTL time.Duration
    secret     []byte
}

func NewJWTManager(secret string, accessTTL, refreshTTL time.Duration) *JWTManager {
    return &JWTManager{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

type AccessClaims struct {
    UserID string   `json:"sub"`
    Roles  []string `json:"roles"`
    jwt.RegisteredClaims
}

type RefreshClaims struct {
    UserID string `json:"sub"`
    jwt.RegisteredClaims
}

func (m *JWTManager) GenerateAccess(userID string, roles []string) (string, time.Time, error) {
    now := time.Now().UTC()
    exp := now.Add(m.accessTTL)
    claims := AccessClaims{UserID: userID, Roles: roles, RegisteredClaims: jwt.RegisteredClaims{Subject: userID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(exp)}}
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    s, err := token.SignedString(m.secret)
    return s, exp, err
}

func (m *JWTManager) GenerateRefresh(userID string) (string, time.Time, error) {
    now := time.Now().UTC()
    exp := now.Add(m.refreshTTL)
    claims := RefreshClaims{UserID: userID, RegisteredClaims: jwt.RegisteredClaims{Subject: userID, IssuedAt: jwt.NewNumericDate(now), ExpiresAt: jwt.NewNumericDate(exp)}}
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    s, err := token.SignedString(m.secret)
    return s, exp, err
}

func (m *JWTManager) ParseAccess(tokenStr string) (*AccessClaims, error) {
    claims := &AccessClaims{}
    t, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) { return m.secret, nil })
    if err != nil || !t.Valid { return nil, ErrInvalidToken }
    return claims, nil
}

func (m *JWTManager) ParseRefresh(tokenStr string) (*RefreshClaims, error) {
    claims := &RefreshClaims{}
    t, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) { return m.secret, nil })
    if err != nil || !t.Valid { return nil, ErrInvalidToken }
    return claims, nil
}
