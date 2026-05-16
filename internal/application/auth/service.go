package auth

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/vishalyadav0987/expense-analyser/internal/domain/auth"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrInvalidOTP         = errors.New("invalid or expired OTP")
)

const (
	MaxFailedAttempts = 3
	FailWindow        = 10 * time.Minute
	LockoutDuration   = 15 * time.Minute
)

// Provider Interfaces:
// By defining these here, the Service is completely decoupled from AWS SES or JWT libraries.
type TokenProvider interface {
	GenerateOTPToken(userID string) (string, error)
	GenerateSessionTokens(userID string) (accessToken string, refreshToken string, err error)
	VerifyToken(tokenString string, expectedAudience string) (string, error)
	VerifyTokenWithClaims(tokenString string, expectedAudience string) (jwt.MapClaims, error)
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
	SetMPIN(ctx context.Context, userID string, mpin string) (string, string, error)
	LoginMPIN(ctx context.Context, email string, mpin string) (string, string, *auth.User, error)

	// ------------------------------------------------------------------
	// The Industry Standard: The "Secure Enclave" Approach
	// ------------------------------------------------------------------
	BiometricLogin(ctx context.Context, refreshToken string) (string, string, *auth.User, error)
}

// Service orchestrates the authentication business logic.
type Service struct {
	userRepo      auth.UserRepository
	otpRepo       auth.OTPRepository
	tokenProvider TokenProvider
	emailProvider EmailProvider
	securityRepo  SecurityRepository
	tokenRepo     auth.TokenRepository
}

// NewService is the dependency injection constructor.
func NewService(
	ur auth.UserRepository,
	or auth.OTPRepository,
	tp TokenProvider,
	ep EmailProvider,
	securityRepo SecurityRepository,
	tokenRepo auth.TokenRepository,
) *Service {
	return &Service{
		userRepo:      ur,
		otpRepo:       or,
		tokenProvider: tp,
		emailProvider: ep,
		securityRepo:  securityRepo,
		tokenRepo:     tokenRepo,
	}
}

// RequestOTP handles both Registration and Login flows.
// It checks if the user exists. If yes -> verify password. If no -> create placeholder account.
func (s *Service) RequestOTP(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
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
			ID:            generateUUID(), // Assume a UUID generator helper exists
			Email:         email,
			PasswordHash:  string(hashedBytes),
			MPINHash:      nil, // Explicitly nil (Not set yet)
			IsActive:      true,
			SetupComplete: false,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
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
	if err != nil {
		// SDE3 Fix: Don't mask database failures! If the DB is down, log it or return a server error.
		return "", false, fmt.Errorf("database error during verification: %w", err)
	}

	if user == nil {
		// SDE3 Fix: If we reach here, it means the OTP was valid, but the user vanished from Postgres.
		// This is a severe state mismatch. We return a clean error to the client.
		return "", false, errors.New("account anomaly: user record missing")
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

// SetMPIN is strictly for the initial setup.
// It requires the userID that was extracted from the otpAccessToken by the middleware.
func (s *Service) SetMPIN(ctx context.Context, userID string, mpin string) (string, string, error) {
	if len(mpin) != 4 {
		return "", "", errors.New("MPIN must be exactly 4 digits")
	}

	// 1. Hash the new MPIN
	hashedMPIN, err := bcrypt.GenerateFromPassword([]byte(mpin), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("failed to hash MPIN: %w", err)
	}

	// 2. Save it to the database
	if err := s.userRepo.UpdateMPIN(ctx, userID, string(hashedMPIN)); err != nil {
		return "", "", fmt.Errorf("failed to save new MPIN: %w", err)
	}

	// 3. Issue the final session tokens
	return s.tokenProvider.GenerateSessionTokens(userID)
}

// LoginMPIN is for everyday quick access.
// It requires NO tokens. It relies on checking the database against the provided email.
func (s *Service) LoginMPIN(ctx context.Context, email string, mpin string) (string, string, *auth.User, error) {
	if len(mpin) != 4 {
		return "", "", &auth.User{}, errors.New("MPIN must be exactly 4 digits")
	}

	// ---------------------------------------------
	//             RATE LIMITTING
	// ---------------------------------------------

	// 1. SECURITY CHECK: Is the account currently locked?
	lockTTL, err := s.securityRepo.GetLockTTL(ctx, email)
	if err != nil {
		return "", "", &auth.User{}, fmt.Errorf("system error checking lock: %w", err)
	}

	if lockTTL > 0 {
		minutesLeft := int(lockTTL.Minutes())
		if minutesLeft == 0 {
			minutesLeft = 1
		} // Show at least 1 min
		return "", "", &auth.User{}, fmt.Errorf("account temporarily locked due to too many failed attempts. Try again in %d minutes", minutesLeft)
	}

	// 1. Find the user by Email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", "", &auth.User{}, fmt.Errorf("database error: %w", err)
	}
	fmt.Println(user)
	if user == nil {
		return "", "", &auth.User{}, errors.New("invalid credentials") // Don't leak that the user doesn't exist
	}

	// 2. Check if they even have an MPIN set up
	if user.MPINHash == nil {
		return "", "", &auth.User{}, errors.New("MPIN not set up for this user. Please verify OTP first.")
	}

	// 3. Verify the MPIN matches the hash in the DB
	if err := bcrypt.CompareHashAndPassword([]byte(*user.MPINHash), []byte(mpin)); err != nil {
		fails, redisErr := s.securityRepo.RecordFailedAttempt(ctx, email, FailWindow)
		if redisErr != nil {
			fmt.Printf("Warning: failed to record auth failure in Redis: %v\n", redisErr)
		} else if fails >= MaxFailedAttempts {
			// Lock them out!
			_ = s.securityRepo.LockAccount(ctx, email, LockoutDuration)
			return "", "", &auth.User{}, errors.New("invalid MPIN. Maximum attempts reached. Account locked for 15 minutes")
		}

		remaining := MaxFailedAttempts - fails
		return "", "", &auth.User{}, fmt.Errorf("invalid MPIN. You have %d attempts remaining", remaining)

	}

	// 4. SUCCESS! Clear all failed attempts and locks.
	_ = s.securityRepo.ClearLockAndAttempts(ctx, email)

	accessToken, refreshToken, err := s.tokenProvider.GenerateSessionTokens(user.ID)

	// 5. Issue fresh session tokens
	return accessToken, refreshToken, user, err
}

// ------------------------------------------------------------------
// The Industry Standard: The "Secure Enclave" Approach
// ------------------------------------------------------------------
func (s *Service) BiometricLogin(ctx context.Context, refreshToken string) (string, string, *auth.User, error) {
	// 1. Verify the Refresh Token mathematically
	claims, err := s.tokenProvider.VerifyTokenWithClaims(refreshToken, "token_refresh")
	if err != nil {
		return "", "", nil, fmt.Errorf("invalid or expired refresh token: %w", err)
	}

	userID, _ := claims["sub"].(string)
	tokenJTI, _ := claims["jti"].(string) // The unique ID of this specific refresh token

	// ---------------------------------------------------------
	// 🚨 DEFENSE LEVEL 1: TOKEN REUSE DETECTION
	// ---------------------------------------------------------
	// Check if this specific token was already used
	isUsed, err := s.tokenRepo.IsTokenUsed(ctx, tokenJTI)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to verify token state: %w", err)
	}
	if isUsed {
		// THEFT DETECTED! Someone is using a token we already rotated.
		// Activate the Kill Switch to destroy ALL sessions for this user.
		err = s.tokenRepo.ActivateKillSwitch(ctx, userID)

		// Force the user out. They must re-authenticate with MPIN.
		return "", "", nil, errors.New("security alert: token compromise detected, please log in with MPIN")
	}

	// ---------------------------------------------------------
	// 🛡️ DEFENSE LEVEL 2: THE KILL SWITCH CHECK
	// ---------------------------------------------------------
	iatFloat, ok := claims["iat"].(float64)
	if !ok {
		return "", "", nil, errors.New("invalid token: missing issued at time")
	}
	issuedAt := time.Unix(int64(iatFloat), 0)

	isKilled, err := s.tokenRepo.IsKillSwitchActive(ctx, userID, issuedAt)
	if err != nil || isKilled {
		return "", "", nil, errors.New("session terminated by security protocol, please log in with MPIN")
	}

	// ---------------------------------------------------------
	// ✅ SUCCESS PATH
	// ---------------------------------------------------------

	// 2. Fetch the User from the database
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", nil, fmt.Errorf("database error: %w", err)
	}

	// 3. Security Guardrails
	if user == nil {
		return "", "", nil, errors.New("access denied: user not found")
	}
	if !user.IsActive {
		// SDE3 Catch: Never let a banned user refresh their session!
		return "", "", nil, errors.New("access denied: account has been deactivated")
	}

	// 4. Generate a brand new Access Token and Refresh Token pair
	newAccessToken, newRefreshToken, err := s.tokenProvider.GenerateSessionTokens(userID)
	if err != nil {
		return "", "", nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	// Safely extract Expiration (exp) to calculate TTL
	expFloat, ok := claims["exp"].(float64)
	if !ok {
		return "", "", nil, errors.New("invalid token: missing expiration time")
	}
	expirationTime := time.Unix(int64(expFloat), 0)
	timeRemaining := time.Until(expirationTime)

	// Blacklist the old token
	if err := s.tokenRepo.MarkTokenUsed(ctx, tokenJTI, timeRemaining); err != nil {
		return "", "", nil, fmt.Errorf("failed to blacklist old token: %w", err)
	}

	// 5. Success! Return the tokens and the user object for the frontend dashboard
	return newAccessToken, newRefreshToken, user, nil
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
