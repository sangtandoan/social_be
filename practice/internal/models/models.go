package models

type User struct {
	Username string `json:"user_name"`
	Email    string `json:"email"`
	ID       int    `json:"id"`
	Age      int    `json:"age"`
}
