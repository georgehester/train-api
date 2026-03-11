package model

type PasswordResetRequest struct {
	Email string `json:"email"`
} // @name PasswordResetRequest

type PasswordUpdateRequest struct {
	Email           string `json:"email"`
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
} // @name PasswordUpdateRequest
