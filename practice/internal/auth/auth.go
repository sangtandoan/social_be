package auth

import (
	"context"
	"crypto/rsa"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthConfig represents the configuration for authentication
type AuthConfig struct {
	// JWT settings
	JWTPrivateKey *rsa.PrivateKey
	JWTPublicKey  *rsa.PublicKey
	TokenExpiry   time.Duration

	// Database connection
	DB *gorm.DB

	// Optional external auth providers
	OAuthProviders map[string]OAuthProvider
}

// User model with extended authentication fields
type User struct {
	gorm.Model
	Email            string `gorm:"unique;not null"`
	PasswordHash     string `gorm:"not null"`
	Role             UserRole
	MFASecret        string
	LastLogin        time.Time
	FailedLoginCount int
	Locked           bool
}

// UserRole represents different user access levels
type UserRole string

const (
	RoleAdmin    UserRole = "ADMIN"
	RoleManager  UserRole = "MANAGER"
	RoleUser     UserRole = "USER"
	RoleReadOnly UserRole = "READ_ONLY"
)

// AuthService handles authentication and authorization
type AuthService struct {
	config *AuthConfig
}

// Credentials for login
type Credentials struct {
	Email    string
	Password string
	MFAToken string
}

// TokenPair represents access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

// OAuthProvider interface for external authentication
type OAuthProvider interface {
	Authenticate(code string) (*User, error)
}

// NewAuthService creates a new authentication service
func NewAuthService(config *AuthConfig) *AuthService {
	return &AuthService{config: config}
}

// Register creates a new user account
func (s *AuthService) Register(
	ctx context.Context,
	email, password string,
	role UserRole,
) (*User, error) {
	// Validate email format
	if !isValidEmail(email) {
		return nil, errors.New("invalid email format")
	}

	// Check if user already exists
	var existingUser User
	result := s.config.DB.Where("email = ?", email).First(&existingUser)
	if result.Error == nil {
		return nil, errors.New("user already exists")
	}

	// Hash password
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("password hashing failed: %v", err)
	}

	// Create new user
	user := &User{
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         role,
	}

	if err := s.config.DB.Create(user).Error; err != nil {
		return nil, fmt.Errorf("user creation failed: %v", err)
	}

	return user, nil
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, creds Credentials) (*TokenPair, error) {
	var user User
	result := s.config.DB.Where("email = ?", creds.Email).First(&user)
	if result.Error != nil {
		return nil, errors.New("user not found")
	}

	// Check if account is locked
	if user.Locked {
		return nil, errors.New("account is locked")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(creds.Password)); err != nil {
		user.FailedLoginCount++
		// Lock account after 5 failed attempts
		if user.FailedLoginCount >= 5 {
			user.Locked = true
		}
		s.config.DB.Save(&user)
		return nil, errors.New("invalid credentials")
	}

	// Reset failed login count
	user.FailedLoginCount = 0
	user.LastLogin = time.Now()
	s.config.DB.Save(&user)

	// Generate tokens
	return s.generateTokenPair(&user)
}

// generateTokenPair creates access and refresh tokens
func (s *AuthService) generateTokenPair(user *User) (*TokenPair, error) {
	// Access Token
	accessToken, err := s.createJWTToken(user, s.config.TokenExpiry)
	if err != nil {
		return nil, err
	}

	// Refresh Token
	refreshToken, err := s.createRefreshToken(user)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// createJWTToken generates a JWT token
func (s *AuthService) createJWTToken(user *User, expiry time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"role":  user.Role,
		"exp":   time.Now().Add(expiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.config.JWTPrivateKey)
}

// createRefreshToken generates a long-lived refresh token
func (s *AuthService) createRefreshToken(user *User) (string, error) {
	token := uuid.New().String()
	// Store refresh token in database with expiration
	// In a real implementation, you'd store this securely
	return token, nil
}

// Middleware for role-based access control
func (s *AuthService) RoleMiddleware(allowedRoles ...UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user role from context (set by authentication middleware)
			userRole, ok := r.Context().Value("role").(UserRole)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Check if user's role is in allowed roles
			for _, role := range allowedRoles {
				if userRole == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, "Insufficient permissions", http.StatusForbidden)
		})
	}
}

// validateToken checks the validity of a JWT token
func (s *AuthService) validateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.config.JWTPublicKey, nil
	})
}

// OAuth Authentication Example
func (s *AuthService) OAuthLogin(provider string, code string) (*TokenPair, error) {
	oauthProvider, exists := s.config.OAuthProviders[provider]
	if !exists {
		return nil, errors.New("provider not supported")
	}

	// Authenticate via external provider
	user, err := oauthProvider.Authenticate(code)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(user)
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	// Implement robust email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}
