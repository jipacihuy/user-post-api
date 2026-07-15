package main

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type Post struct {
	ID      int    `json:"id"`
	Title   string `json:"title" validate:"required,min=3,max=200"`
	Content string `json:"content" validate:"required,min=1"`
	UserID  int    `json:"user_id" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}