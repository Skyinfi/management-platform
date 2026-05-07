package model

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt string    `json:"expiresAt"`
	User      LoginUser `json:"user"`
}

type MeResponse = LoginResponse
