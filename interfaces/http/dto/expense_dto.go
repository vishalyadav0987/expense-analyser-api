package dto

type CreateExpenseRequest struct {
	Name string `json:"name" binding:"required"`
	Type string `json:"type" binding:"required,oneof=Need Want Saving"`
}
