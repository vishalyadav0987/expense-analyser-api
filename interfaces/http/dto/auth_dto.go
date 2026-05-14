package dto

// ------------------------------------------------------------------
// 1. Request OTP (Login / Register)
// ------------------------------------------------------------------

type RequestOTPRequest struct {
	// The `validate` tags tell our middleware to reject bad data before it hits the Service layer.
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RequestOTPResponse struct {
	UserID string `json:"userId"`
}

// ------------------------------------------------------------------
// 2. Verify OTP
// ------------------------------------------------------------------

type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=4,numeric"`
}

type VerifyOTPResponse struct {
	OTPAccessToken string `json:"otpAccessToken"`
	IsNewUser      bool   `json:"isNewUser"`
}

// ------------------------------------------------------------------
// 3. Submit MPIN
// ------------------------------------------------------------------

type SubmitMPINRequest struct {
	// Notice UserID is NOT here.
	// As an SDE3 security practice, we extract UserID from the JWT Token in the headers,
	// NOT from the JSON body, to prevent a user from setting an MPIN for someone else.
	MPIN string `json:"mpin" validate:"required,len=4,numeric"`
}

type SubmitMPINResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}
