package dto

type StaffCreateRequest struct {
	Username   string `json:"username" validate:"required,min=3"`
	Password   string `json:"password" validate:"required,min=8"`
	HospitalID int    `json:"hospital_id" validate:"required,gt=0"`
}

type StaffCreateResponse struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	HospitalID int    `json:"hospital_id"`
}

// StaffLoginRequest represents the payload for staff login.
// Username and password are required fields validated by the handler.
type StaffLoginRequest struct {
	Username string `json:"username" validate:"required,min=3"`
	Password string `json:"password" validate:"required,min=8"`
}

// StaffLoginResponse contains the authentication tokens returned after
// a successful login.
type StaffLoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"` // seconds until expiration
}
