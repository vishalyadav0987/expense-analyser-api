package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/vishalyadav0987/expense-analyser/internal/domain/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidOTP         = errors.New("invalid or expired OTP")
)

// Provider Interfaces:
// By defining these here, the Service is completely decoupled from AWS SES or JWT libraries.
type TokenProvider interface {
	GenerateOTPToken(userID string) (string, error)
	GenerateSessionTokens(userID string) (accessToken string, refreshToken string, err error)
}

type EmailProvider interface {
	SendOTP(ctx context.Context, email, otp string) error
}

// ------------------------------------------------------------------
// THE FIX: The UseCase Interface
// This is the contract the HTTP Delivery layer will depend on.
// ------------------------------------------------------------------
type UseCase interface {
	RequestOTP(ctx context.Context, email, password string) (string, error)
	VerifyOTP(ctx context.Context, email, otp string) (otpToken string, isNewUser bool, err error)
	SubmitMPIN(ctx context.Context, userID, mpin string) (accessToken string, refreshToken string, err error)
}

// Service orchestrates the authentication business logic.
type Service struct {
	userRepo      auth.UserRepository
	otpRepo       auth.OTPRepository
	tokenProvider TokenProvider
	emailProvider EmailProvider
}

// NewService is the dependency injection constructor.
func NewService(ur auth.UserRepository, or auth.OTPRepository, tp TokenProvider, ep EmailProvider) *Service {
	return &Service{
		userRepo:      ur,
		otpRepo:       or,
		tokenProvider: tp,
		emailProvider: ep,
	}
}

// RequestOTP handles both Registration and Login flows.
// It checks if the user exists. If yes -> verify password. If no -> create placeholder account.
func (s *Service) RequestOTP(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		fmt.Println("pass")

		return "", fmt.Errorf("auth service - failed to fetch user: %w", err)
	}

	var userID string

	if user == nil {
		// New User Registration Flow
		// SDE3 Note: DefaultCost is currently 10. For highly sensitive apps, use 12,
		// but be aware it increases CPU load significantly during login spikes.
		hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return "", fmt.Errorf("failed to hash password: %w", err)
		}

		newUser := &auth.User{
			ID:           generateUUID(), // Assume a UUID generator helper exists
			Email:        email,
			PasswordHash: string(hashedBytes),
			MPINHash:     nil, // Explicitly nil (Not set yet)
			IsActive:     true,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := s.userRepo.CreateUser(ctx, newUser); err != nil {
			return "", fmt.Errorf("failed to register user: %w", err)
		}
		userID = newUser.ID

	} else {
		// Existing User Login Flow
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			// SDE3 Security Guardrail: Never return "password incorrect".
			// Always return generic "invalid credentials" to prevent email enumeration.
			return "", ErrInvalidCredentials
		}
		userID = user.ID
	}

	// Generate a Cryptographically Secure 4-digit OTP
	otp, err := generateSecureOTP()
	if err != nil {
		return "", fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Save to Redis with a 5-minute TTL
	if err := s.otpRepo.SaveOTP(ctx, email, otp, 5*time.Minute); err != nil {
		return "", fmt.Errorf("failed to save OTP: %w", err)
	}

	// Fire and forget the email (In production, this should be pushed to an async message queue like RabbitMQ)
	go func() {
		// Create a detached context so the email still sends even if the HTTP request finishes
		bgCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = s.emailProvider.SendOTP(bgCtx, email, otp)
	}()

	return userID, nil
}

// VerifyOTP validates the OTP and issues a short-lived token to access the MPIN screen.
func (s *Service) VerifyOTP(ctx context.Context, email, otp string) (otpToken string, isNewUser bool, err error) {
	savedOTP, err := s.otpRepo.GetOTP(ctx, email)
	if err != nil || savedOTP != otp {
		return "", false, ErrInvalidOTP
	}

	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil || user == nil {
		return "", false, errors.New("user not found after OTP verification")
	}

	// OTP is single-use. Delete it immediately to prevent replay attacks.
	_ = s.otpRepo.DeleteOTP(ctx, email)

	// Generate a 5-minute token ONLY valid for setting/verifying the MPIN
	otpToken, err = s.tokenProvider.GenerateOTPToken(user.ID)
	if err != nil {
		return "", false, fmt.Errorf("failed to generate OTP token: %w", err)
	}

	return otpToken, !user.HasSetupMPIN(), nil
}

// SubmitMPIN handles both setting a new MPIN and verifying an existing one.
func (s *Service) SubmitMPIN(ctx context.Context, userID, mpin string) (string, string, error) {
	if len(mpin) != 4 {
		return "", "", auth.ErrMPINTooShort
	}

	user, err := s.userRepo.GetUserByEmail(ctx, userID) // Assuming you overload GetUser or use GetUserByID
	if err != nil || user == nil {
		return "", "", errors.New("user not found")
	}

	if !user.HasSetupMPIN() {
		// Setup MPIN Flow
		hashedMPIN, _ := bcrypt.GenerateFromPassword([]byte(mpin), bcrypt.DefaultCost)
		if err := s.userRepo.UpdateMPIN(ctx, user.ID, string(hashedMPIN)); err != nil {
			return "", "", fmt.Errorf("failed to save new MPIN: %w", err)
		}
	} else {
		// Verify MPIN Flow
		if err := bcrypt.CompareHashAndPassword([]byte(*user.MPINHash), []byte(mpin)); err != nil {
			return "", "", errors.New("invalid MPIN")
		}
	}

	// Issue final production tokens
	accessToken, refreshToken, err := s.tokenProvider.GenerateSessionTokens(user.ID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate session tokens: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateSecureOTP creates a random 4 digit string using crypto/rand
func generateSecureOTP() (string, error) {
	max := big.NewInt(10000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}

// Mock UUID generator for example purposes
func generateUUID() string {
	return "usr_" + fmt.Sprintf("%d", time.Now().UnixNano())
}
