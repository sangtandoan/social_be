package dto

type CreateUserRequest struct {
	Username string `json:"username,omitempty" validate:"required,min=3,max=50"`
	Email    string `json:"email,omitempty"    validate:"required,email"`
	Age      int    `json:"age,omitempty"      validate:"required,gte=18"`
}
