package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vishalyadav0987/expense-analyser/interfaces/http/dto"
	"github.com/vishalyadav0987/expense-analyser/internal/application/auth"
)

type AuthHandler struct {
	authService auth.UseCase
}

func NewAuthHandler(svc auth.UseCase) *AuthHandler {
	return &AuthHandler{authService: svc}
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
func (h *AuthHandler) HandleSubmitMPIN(c *gin.Context) {
	var req dto.SubmitMPINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.NewErrorResponse("Invalid payload: "+err.Error()))
		return
	}

	// SDE3 Note: We extract the userID from Gin's context (set by the JWT middleware)
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, dto.NewErrorResponse("Unauthorized: Missing user context"))
		return
	}

	accessToken, refreshToken, err := h.authService.SubmitMPIN(c.Request.Context(), userIDVal.(string), req.MPIN)
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
