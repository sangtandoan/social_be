package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// RefreshToken represents the database model for storing refresh tokens
type RefreshToken struct {
	gorm.Model
	UserID    uint      `gorm:"not null"`
	User      User      `gorm:"foreignkey:UserID"`
	Token     string    `gorm:"unique;not null"`
	ExpiresAt time.Time `gorm:"not null"`
	LastUsed  time.Time
	IPAddress string
	UserAgent string
	IsRevoked bool `gorm:"default:false"`
}

// RefreshTokenService handles refresh token operations
type RefreshTokenService struct {
	db *gorm.DB
}

// RefreshTokenRequest represents the input for refresh token operation
type RefreshTokenRequest struct {
	RefreshToken string
	IPAddress    string
	UserAgent    string
}

// NewRefreshTokenService creates a new service for managing refresh tokens
func NewRefreshTokenService(db *gorm.DB) *RefreshTokenService {
	return &RefreshTokenService{db: db}
}

// GenerateRefreshToken creates a new refresh token for a user
func (s *RefreshTokenService) GenerateRefreshToken(
	ctx context.Context,
	userID uint,
) (*RefreshToken, error) {
	// Generate cryptographically secure random token
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %v", err)
	}
	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)

	// Create refresh token record
	refreshToken := &RefreshToken{
		UserID:    userID,
		Token:     tokenString,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days validity
	}

	// Save to database
	result := s.db.WithContext(ctx).Create(refreshToken)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to save refresh token: %v", result.Error)
	}

	return refreshToken, nil
}

// ValidateRefreshToken checks if a refresh token is valid and returns the associated user
func (s *RefreshTokenService) ValidateRefreshToken(
	ctx context.Context,
	req RefreshTokenRequest,
) (*User, error) {
	var refreshToken RefreshToken

	// Find the refresh token
	result := s.db.WithContext(ctx).
		Preload("User").
		Where("token = ? AND is_revoked = ?", req.RefreshToken, false).
		First(&refreshToken)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid or expired refresh token")
		}
		return nil, result.Error
	}

	// Check if token is expired
	if time.Now().After(refreshToken.ExpiresAt) {
		// Automatically revoke expired token
		s.db.Model(&refreshToken).Update("is_revoked", true)
		return nil, errors.New("refresh token has expired")
	}

	// Validate additional token metadata (optional but recommended)
	if refreshToken.IPAddress != "" && refreshToken.IPAddress != req.IPAddress {
		return nil, errors.New("refresh token used from different IP")
	}

	// Optional: Check user agent for additional security
	if refreshToken.UserAgent != "" && refreshToken.UserAgent != req.UserAgent {
		return nil, errors.New("potential unauthorized access")
	}

	// Update token metadata
	refreshToken.LastUsed = time.Now()
	refreshToken.IPAddress = req.IPAddress
	refreshToken.UserAgent = req.UserAgent
	s.db.Save(&refreshToken)

	return &refreshToken.User, nil
}

// RevokeRefreshToken invalidates a specific refresh token
func (s *RefreshTokenService) RevokeRefreshToken(ctx context.Context, tokenString string) error {
	result := s.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("token = ?", tokenString).
		Update("is_revoked", true)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke token: %v", result.Error)
	}

	return nil
}

// RevokeAllUserTokens invalidates all refresh tokens for a user
func (s *RefreshTokenService) RevokeAllUserTokens(ctx context.Context, userID uint) error {
	result := s.db.WithContext(ctx).
		Model(&RefreshToken{}).
		Where("user_id = ?", userID).
		Update("is_revoked", true)

	if result.Error != nil {
		return fmt.Errorf("failed to revoke user tokens: %v", result.Error)
	}

	return nil
}

// CleanupExpiredTokens removes expired and revoked refresh tokens
func (s *RefreshTokenService) CleanupExpiredTokens(ctx context.Context) error {
	result := s.db.WithContext(ctx).
		Where("expires_at < ? OR is_revoked = ?", time.Now(), true).
		Delete(&RefreshToken{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup tokens: %v", result.Error)
	}

	return nil
}

// ExtendAuthService with refresh token functionality
func (s *AuthService) RefreshAccessToken(
	ctx context.Context,
	req RefreshTokenRequest,
) (*TokenPair, error) {
	// Validate refresh token and get user
	user, err := s.refreshTokenService.ValidateRefreshToken(ctx, req)
	if err != nil {
		return nil, err
	}

	// Generate new token pair
	return s.generateTokenPair(user)
}

// Periodic cleanup job (can be run as a background task)
func (s *RefreshTokenService) StartCleanupJob(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			err := s.CleanupExpiredTokens(context.Background())
			if err != nil {
				// Log error - in a real app, use proper logging
				fmt.Printf("Token cleanup failed: %v\n", err)
			}
		}
	}()
}
