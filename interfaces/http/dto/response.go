package dto

// APIResponse is the universal wrapper for all HTTP responses.
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	// omitempty ensures that if Data is nil, the "data" key doesn't appear as "data": null in the JSON.
	Data interface{} `json:"data,omitempty"`
}

// ErrorResponse is useful for standardizing 4xx and 5xx errors.
// Sometimes you might want to return a list of specific field validation errors.
type ErrorResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Errors  []string `json:"errors,omitempty"` // Example: ["email is required", "password too short"]
}

// NewSuccessResponse is a helper to quickly generate a 200 OK payload
func NewSuccessResponse(message string, data interface{}) APIResponse {
	return APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// NewErrorResponse is a helper to quickly generate a failure payload
func NewErrorResponse(message string, errs ...string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: message,
		Errors:  errs,
	}
}
