package dto

type CreateUserRequest struct {
	Username string `json:"username,omitempty" validate:"required,min=3,max=50"`
	Email    string `json:"email,omitempty"    validate:"required,email"`
	Password string `json:"password,omitempty" validate:"required,min=3,max=20"`
}

type CreateUserResponse struct {
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Token    string `json:"token,omitempty"`
	UserID   int64  `json:"user_id,omitempty"`
}

type ActivateUserRequest struct {
	Token  string
	UserID int64
}

type LoginRequest struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}
