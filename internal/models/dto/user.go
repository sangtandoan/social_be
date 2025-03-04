package dto

type CreateUserRequest struct {
	Username string `json:"username,omitempty" validate:"required,min=3,max=50"`
	Email    string `json:"email,omitempty"    validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required,min=3,max=20"`
}
