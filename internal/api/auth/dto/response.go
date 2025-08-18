package dto

type RegisterResponse struct {
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	UserDetails  *UserDetails `json:"user_details,omitempty"`
	Message      string       `json:"message"`
}

type UserDetails struct {
	Username      string  `json:"username,omitempty"`
	Email         string  `json:"email,omitempty"`
	FullName      string  `json:"full_name,omitempty"`
	ProfilePicURL *string `json:"profile_pic_url,omitempty"`
}

type LoginResponse struct { // TOdo: uniffy with register response and mapper
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	UserDetails  *UserDetails `json:"user_details,omitempty"`
	Message      string       `json:"message"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Message      string `json:"message"`
}
