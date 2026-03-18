package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/scaliann/withdrawal_service_test/internal/domain"
	"github.com/scaliann/withdrawal_service_test/internal/dto"
	"github.com/scaliann/withdrawal_service_test/pkg/otel/tracer"
)

const (
	refreshJTIBytes = 32

	refreshRedisPrefix = "auth:refresh:"

	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"

	jwtAlgorithm = "HS256"
	jwtType      = "JWT"
)

func (u *UseCase) IssueToken(ctx context.Context, input dto.IssueTokenInput) (dto.TokenPairOutput, error) {
	ctx, span := tracer.Start(ctx, "IssueToken")
	defer span.End()

	var output dto.TokenPairOutput

	username := strings.TrimSpace(input.Username)
	password := strings.TrimSpace(input.Password)
	if username == "" || password == "" {
		return output, domain.ErrInvalidCredentials
	}

	user, err := u.postgres.GetUserByCredentials(ctx, username, password)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return output, domain.ErrInvalidCredentials
		}

		return output, fmt.Errorf("postgres.GetUserByCredentials: %w", err)
	}
	if !user.IsActive {
		return output, domain.ErrInvalidCredentials
	}

	output, err = u.issueTokenPair(ctx, user.Username, user.Role, user.ID)
	if err != nil {
		return output, fmt.Errorf("issueTokenPair: %w", err)
	}

	return output, nil
}

func (u *UseCase) RefreshToken(ctx context.Context, input dto.RefreshTokenInput) (dto.TokenPairOutput, error) {
	ctx, span := tracer.Start(ctx, "RefreshToken")
	defer span.End()

	var output dto.TokenPairOutput

	refreshToken := strings.TrimSpace(input.RefreshToken)
	if refreshToken == "" {
		return output, domain.ErrRefreshTokenRequired
	}

	claims, err := u.parseAndValidateToken(refreshToken, tokenTypeRefresh)
	if err != nil {
		return output, fmt.Errorf("parseAndValidateToken: %w", err)
	}

	refreshKey := refreshRedisKey(claims.JWTID)
	data, err := u.redis.Get(ctx, refreshKey)
	if err != nil {
		if errors.Is(err, domain.ErrKeyNotFound) {
			return output, domain.ErrUnauthorized
		}

		return output, fmt.Errorf("redis.Get: %w", err)
	}

	subject := strings.TrimSpace(string(data))
	if subject == "" || !secureEqual(subject, claims.Subject) {
		return output, domain.ErrUnauthorized
	}

	err = u.redis.Del(ctx, refreshKey)
	if err != nil {
		if errors.Is(err, domain.ErrKeyNotFound) {
			return output, domain.ErrUnauthorized
		}

		return output, fmt.Errorf("redis.Del: %w", err)
	}

	output, err = u.issueTokenPair(ctx, claims.Subject, claims.Role, claims.UserID)
	if err != nil {
		return output, fmt.Errorf("issueTokenPair: %w", err)
	}

	return output, nil
}

func (u *UseCase) VerifyAccessToken(ctx context.Context, accessToken string) (JwtClaims, error) {
	ctx, span := tracer.Start(ctx, "VerifyAccessToken")
	defer span.End()

	token := strings.TrimSpace(accessToken)
	if token == "" {
		return JwtClaims{}, domain.ErrAccessTokenRequired
	}

	claims, err := u.parseAndValidateToken(token, tokenTypeAccess)
	if err != nil {
		return JwtClaims{}, fmt.Errorf("parseAndValidateToken: %w", err)
	}

	return claims, nil
}

func (u *UseCase) issueTokenPair(ctx context.Context, subject string, role string, userID int) (dto.TokenPairOutput, error) {
	var output dto.TokenPairOutput

	subject = strings.TrimSpace(subject)
	if subject == "" {
		return output, domain.ErrUnauthorized
	}

	accessToken, err := u.buildJWT(subject, role, userID, tokenTypeAccess, "", u.auth.AccessTTL)
	if err != nil {
		return output, fmt.Errorf("build access token: %w", err)
	}

	refreshJTI, err := generateToken(refreshJTIBytes)
	if err != nil {
		return output, fmt.Errorf("generate refresh jti: %w", err)
	}

	refreshToken, err := u.buildJWT(subject, role, userID, tokenTypeRefresh, refreshJTI, u.auth.RefreshTTL)
	if err != nil {
		return output, fmt.Errorf("build refresh token: %w", err)
	}

	err = u.redis.Set(ctx, refreshRedisKey(refreshJTI), []byte(subject), u.auth.RefreshTTL)
	if err != nil {
		return output, fmt.Errorf("redis.Set refresh token: %w", err)
	}

	output = dto.TokenPairOutput{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(u.auth.AccessTTL.Seconds()),
	}

	return output, nil
}

func (u *UseCase) parseAndValidateToken(token string, expectedTokenType string) (JwtClaims, error) {
	claims, err := parseAndVerifyJWT(token, u.auth.JWTSecret)
	if err != nil {
		return JwtClaims{}, domain.ErrUnauthorized
	}

	err = u.validateClaims(claims, expectedTokenType)
	if err != nil {
		return JwtClaims{}, err
	}

	return claims, nil
}

func (u *UseCase) buildJWT(subject string, role string, userID int, tokenType string, jti string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(u.auth.JWTSecret) == "" {
		return "", fmt.Errorf("empty jwt secret")
	}

	now := time.Now().UTC()
	claims := JwtClaims{
		Issuer:    strings.TrimSpace(u.auth.JWTIssuer),
		Subject:   strings.TrimSpace(subject),
		Role:      strings.TrimSpace(role),
		UserID:    userID,
		ExpiresAt: now.Add(ttl).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		TokenType: tokenType,
		JWTID:     strings.TrimSpace(jti),
	}

	headerPart, err := encodeJWTPart(jwtHeader{
		Algorithm: jwtAlgorithm,
		Type:      jwtType,
	})
	if err != nil {
		return "", fmt.Errorf("encode header: %w", err)
	}

	claimsPart, err := encodeJWTPart(claims)
	if err != nil {
		return "", fmt.Errorf("encode claims: %w", err)
	}

	signingInput := headerPart + "." + claimsPart
	signature := signHS256(signingInput, u.auth.JWTSecret)

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature), nil
}

func (u *UseCase) validateClaims(claims JwtClaims, expectedTokenType string) error {
	nowUnix := time.Now().UTC().Unix()

	if strings.TrimSpace(claims.Subject) == "" {
		return domain.ErrUnauthorized
	}

	if claims.TokenType != expectedTokenType {
		return domain.ErrUnauthorized
	}

	if claims.ExpiresAt == 0 || nowUnix >= claims.ExpiresAt {
		return domain.ErrUnauthorized
	}

	if claims.NotBefore != 0 && nowUnix < claims.NotBefore {
		return domain.ErrUnauthorized
	}

	if claims.IssuedAt != 0 && claims.IssuedAt > nowUnix {
		return domain.ErrUnauthorized
	}

	expectedIssuer := strings.TrimSpace(u.auth.JWTIssuer)
	if expectedIssuer != "" && claims.Issuer != expectedIssuer {
		return domain.ErrUnauthorized
	}

	if expectedTokenType == tokenTypeRefresh && strings.TrimSpace(claims.JWTID) == "" {
		return domain.ErrUnauthorized
	}

	return nil
}

func parseAndVerifyJWT(token string, jwtSecret string) (JwtClaims, error) {
	if strings.TrimSpace(jwtSecret) == "" {
		return JwtClaims{}, fmt.Errorf("empty jwt secret")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return JwtClaims{}, fmt.Errorf("invalid token parts")
	}

	var header jwtHeader
	err := decodeJWTPart(parts[0], &header)
	if err != nil {
		return JwtClaims{}, fmt.Errorf("decode header: %w", err)
	}
	if header.Algorithm != jwtAlgorithm || header.Type != jwtType {
		return JwtClaims{}, fmt.Errorf("unexpected header")
	}

	signingInput := parts[0] + "." + parts[1]
	gotSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return JwtClaims{}, fmt.Errorf("decode signature: %w", err)
	}

	expectedSignature := signHS256(signingInput, jwtSecret)
	if !hmac.Equal(gotSignature, expectedSignature) {
		return JwtClaims{}, fmt.Errorf("invalid signature")
	}

	var claims JwtClaims
	err = decodeJWTPart(parts[1], &claims)
	if err != nil {
		return JwtClaims{}, fmt.Errorf("decode claims: %w", err)
	}

	return claims, nil
}

func generateToken(size int) (string, error) {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", fmt.Errorf("rand.Read: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func refreshRedisKey(jti string) string {
	return refreshRedisPrefix + tokenHash(jti)
}

func tokenHash(value string) string {
	hash := sha256.Sum256([]byte(value))

	return hex.EncodeToString(hash[:])
}

func signHS256(payload string, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(payload))

	return mac.Sum(nil)
}

func encodeJWTPart(v any) (string, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("json.Marshal: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func decodeJWTPart(part string, v any) error {
	payload, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return fmt.Errorf("base64 decode: %w", err)
	}

	err = json.Unmarshal(payload, v)
	if err != nil {
		return fmt.Errorf("json.Unmarshal: %w", err)
	}

	return nil
}

func secureEqual(got string, expected string) bool {
	if len(got) != len(expected) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(got), []byte(expected)) == 1
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

type JwtClaims struct {
	Issuer    string `json:"iss,omitempty"`
	Subject   string `json:"sub"`
	UserID    int    `json:"user_id"`
	Role      string `json:"role"`
	ExpiresAt int64  `json:"exp"`
	IssuedAt  int64  `json:"iat"`
	NotBefore int64  `json:"nbf"`
	TokenType string `json:"token_type"`
	JWTID     string `json:"jti,omitempty"`
}
