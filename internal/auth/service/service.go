package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/undndnwnkk/go-warehouse-system/internal/auth/model"
	"github.com/undndnwnkk/go-warehouse-system/internal/auth/repository"
	"github.com/undndnwnkk/go-warehouse-system/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"net/mail"
	"time"
)

type AuthService struct {
	userRepository  repository.PostgresUserRepository
	tokenRepository repository.PostgresRefreshTokenRepository
	jwtManager      *jwt.Manager
	accessTTL       time.Duration
	refreshTTL      time.Duration
}

func NewAuthService(
	userRepo repository.PostgresUserRepository,
	tokenRepo repository.PostgresRefreshTokenRepository,
	jwtManager *jwt.Manager,
	accessTTL time.Duration,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		userRepository:  userRepo,
		tokenRepository: tokenRepo,
		jwtManager:      jwtManager,
		accessTTL:       accessTTL,
		refreshTTL:      refreshTTL,
	}
}

func (s *AuthService) Register(ctx context.Context, req model.CreateUserRequest) (*model.TokenPair, error) {
	if !validEmail(req.Email) {
		return nil, fmt.Errorf("invalid email")
	}

	if len(req.Password) < 8 {
		return nil, fmt.Errorf("invalid password")
	}

	u, err := s.userRepository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if u != nil {
		return nil, fmt.Errorf("incorrect email or password")
	}

	password_hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("error while hashing: %w", err)
	}

	currentUser := model.User{Email: req.Email, PasswordHash: password_hash}

	id, err := s.userRepository.SaveUser(ctx, currentUser)
	if err != nil {
		return nil, fmt.Errorf("error while saving user: %w", err)
	}

	return s.generateTokenPair(ctx, id)
}

func (s *AuthService) Login(ctx context.Context, req model.CreateUserRequest) (*model.TokenPair, error) {
	u, err := s.userRepository.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if !CheckPasswordHash(req.Password, u.PasswordHash) {
		return nil, fmt.Errorf("invalid email or password")
	}

	return s.generateTokenPair(ctx, u.ID)
}

func (s *AuthService) Refresh(ctx context.Context, userID, rawRefreshToken string) (*model.TokenPair, error) {
	if rawRefreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	activeTokens, err := s.tokenRepository.GetActiveRefreshTokensByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user tokens: %w", err)
	}

	var matchedToken *model.RefreshToken

	for _, t := range activeTokens {
		if CheckPasswordHash(rawRefreshToken, t.TokenHash) {
			matchedToken = &t
			break
		}
	}

	if matchedToken == nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	if matchedToken.Revoked {
		return nil, fmt.Errorf("refresh token has been revoked")
	}

	if time.Now().After(matchedToken.ExpiresAt) {
		return nil, fmt.Errorf("refresh token has expired")
	}

	if err := s.tokenRepository.RevokeRefreshToken(ctx, matchedToken.ID); err != nil {
		return nil, fmt.Errorf("failed to revoke old refresh token: %w", err)
	}

	newTokenPair, err := s.generateTokenPair(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new token pair: %w", err)
	}

	return newTokenPair, nil
}

func (s *AuthService) generateTokenPair(ctx context.Context, userID string) (*model.TokenPair, error) {
	accessToken, err := s.jwtManager.GenerateToken(userID, s.accessTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	rawRefreshToken, err := generateRandomString(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token string: %w", err)
	}

	refreshTokenHash, err := HashPassword(rawRefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %w", err)
	}

	refreshTokenModel := model.RefreshToken{
		UserID:    userID,
		TokenHash: refreshTokenHash,
		Revoked:   false,
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}

	if err := s.tokenRepository.SaveRefreshToken(ctx, refreshTokenModel); err != nil {
		return nil, fmt.Errorf("failed to save refresh token to database: %w", err)
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
	}, nil
}

func generateRandomString(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
