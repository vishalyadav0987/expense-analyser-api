package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/vishalyadav0987/expense-analyser/interfaces/http/dto"
	"github.com/vishalyadav0987/expense-analyser/internal/application/auth"
)

type AuthHandler struct {
	authService  auth.UseCase
	tokenManager auth.TokenProvider
}

func NewAuthHandler(svc auth.UseCase, tokenManager auth.TokenProvider) *AuthHandler {
	return &AuthHandler{authService: svc, tokenManager: tokenManager}
}

// HandleRequestOTP maps to POST /api/v1/auth/request-otp
func (h *AuthHandler) HandleRequestOTP(c *gin.Context) {
	var req dto.RequestOTPRequest

	// SDE3 Magic: ShouldBindJSON automatically checks the 'binding' tags in your DTO.
	// If the email is missing or invalid, it immediately returns an error. No manual checks needed.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid payload: "+err.Error()))
		return
	}

	userID, err := h.authService.RequestOTP(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Invalid credentials or system error"))
		return
	}

	respData := dto.RequestOTPResponse{UserID: userID}
	c.JSON(http.StatusOK, dto.NewSuccessResponse("OTP sent successfully.", respData))
}

// HandleVerifyOTP maps to POST /api/v1/auth/verify-otp
func (h *AuthHandler) HandleVerifyOTP(c *gin.Context) {
	var req dto.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid payload: "+err.Error()))
		return
	}

	otpToken, isNewUser, err := h.authService.VerifyOTP(c.Request.Context(), req.Email, req.OTP)
	if err != nil {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Invalid or expired OTP"))
		return
	}

	respData := dto.VerifyOTPResponse{
		OTPAccessToken: otpToken,
		IsNewUser:      isNewUser,
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse("OTP verified successfully.", respData))
}

// HandleSubmitMPIN maps to POST /api/v1/auth/submit-mpin
func (h *AuthHandler) HandleSetMPIN(c *gin.Context) {
	var req dto.SubmitMPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid payload: "+err.Error()))
		return
	}

	// SDE3 Note: We extract the userID from Gin's context (set by the JWT middleware)
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse("Authorization header required"))
		return
	}

	// 2. Format check: It must look like "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse("Invalid authorization format. Expected 'Bearer <token>'"))
		return
	}

	tokenString := parts[1]

	// 3. Verify the Token using the provider we just updated
	userIDVal, err := h.tokenManager.VerifyToken(tokenString, "mpin_setup")
	fmt.Println(userIDVal)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, dto.NewErrorResponse("Invalid or expired token: "+err.Error()))
		return
	}

	accessToken, refreshToken, err := h.authService.SetMPIN(c.Request.Context(), userIDVal, req.MPIN)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse(err.Error()))
		return
	}

	respData := dto.SubmitMPINResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse("MPIN processed successfully.", respData))
}

// HandleLoginMPIN maps to POST /api/v1/auth/login-mpin
// This is an UNPROTECTED route. It requires no JWT headers.
func (h *AuthHandler) HandleLoginMPIN(c *gin.Context) {
	var req dto.LoginMPINRequest

	// 1. Validate the incoming JSON body (Email and 4-digit MPIN)
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid payload: "+err.Error()))
		return
	}

	// 2. Call the Service layer to verify the credentials
	accessToken, refreshToken, user, err := h.authService.LoginMPIN(c.Request.Context(), req.Email, req.MPIN)
	if err != nil {
		// If the MPIN is wrong or the user doesn't exist, we return a 401 Unauthorized
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse(err.Error()))
		return
	}

	// 3. Success! Return the session tokens
	respData := dto.LoginMPINResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserData: dto.UserDataRes{
			Email:         req.Email,
			UserID:        user.ID,
			Active:        user.IsActive,
			SetupComplete: user.SetupComplete,
		},
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Login successful", respData))
}

func (h *AuthHandler) HandleBiometricLogin(c *gin.Context) {
	// 1. Parse and Validate the JSON payload
	var req dto.BiometricLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("invalid payload: refresh token is required"))
		return
	}

	// 2. Call the Biometric Login Service
	accessToken, newRefreshToken, user, err := h.authService.BiometricLogin(c.Request.Context(), req.RefreshToken)
	if err != nil {
		errStr := err.Error()

		// SDE3 Tip: Log the raw error internally for your backend monitoring (Datadog, Kibana, etc.)
		// log.Printf("[AUTH ERROR] Biometric login failed: %v", err)

		// 🚨 CASE 1: Active Security Threats (Kill Switch / Token Theft)
		if strings.Contains(errStr, "security alert") || strings.Contains(errStr, "session terminated") {
			// Frontend should instantly wipe the local vault and show the MPIN screen
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Security alert: session terminated to protect your account. Please log in with MPIN."))
			return
		}

		// ⚠️ CASE 2: Token Validation Failures (Expired, Malformed, Missing Claims)
		if strings.Contains(errStr, "invalid") || strings.Contains(errStr, "expired") {
			c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Session expired or invalid. Please log in with MPIN again."))
			return
		}

		// 🚫 CASE 3: Account Status Issues (Banned or Deleted)
		if strings.Contains(errStr, "deactivated") || strings.Contains(errStr, "not found") {
			c.JSON(http.StatusForbidden, dto.NewErrorResponse("Access denied: Account is deactivated or unavailable."))
			return
		}

		// 🛠️ CASE 4: The New Fallback (Database down, Redis timeout, etc.)
		// We use 500 here because it's our fault, not the user's.
		// We don't force them to use MPIN immediately, we just tell them to try again.
		c.JSON(http.StatusInternalServerError, dto.NewErrorResponse("An unexpected server error occurred. Please try again later."))
		return
	}

	// 3. Return the exact same successful login response format as your MPIN login!
	respData := dto.LoginMPINResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		UserData: dto.UserDataRes{
			Email:         user.Email,
			UserID:        user.ID,
			Active:        user.IsActive,
			SetupComplete: user.SetupComplete,
		},
	}
	c.JSON(http.StatusOK, dto.NewSuccessResponse("Biometric login successful.", respData))
}
